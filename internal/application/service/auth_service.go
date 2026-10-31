package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	serviceinterface "github.com/tahedig/too-good-to-go/backend/internal/application/service/interface"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/enum"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/exception"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/model"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/repository"
	"github.com/tahedig/too-good-to-go/backend/pkg/logger"
	"github.com/tahedig/too-good-to-go/backend/pkg/security"
	"github.com/tahedig/too-good-to-go/backend/pkg/validation"
)

// OTPConfig controls OTP generation and verification.
type OTPConfig struct {
	Length      int
	TTL         time.Duration
	MaxAttempts int
	// ResendCooldown is the minimum time between two codes for the same
	// phone. It limits SMS cost and abuse; zero disables it.
	ResendCooldown time.Duration
	// ExposeCode returns the plain code to the caller. Only for testing
	// without real SMS delivery; config refuses it in production.
	ExposeCode bool
}

type authService struct {
	repo   repository.OTPRepository
	users  repository.UserRepository
	sms    serviceinterface.SMSSender
	tokens serviceinterface.TokenManager
	cfg    OTPConfig
	roles  map[string]enum.Role
	log    *slog.Logger
	now    func() time.Time
}

// NewAuthService wires the OTP + login flow. phoneRoles maps normalized phone
// numbers to the role a brand-new user with that number receives; everyone
// else becomes a customer. Roles of existing users are never changed here.
func NewAuthService(
	repo repository.OTPRepository,
	users repository.UserRepository,
	sms serviceinterface.SMSSender,
	tokens serviceinterface.TokenManager,
	cfg OTPConfig,
	phoneRoles map[string]enum.Role,
	log *slog.Logger,
) serviceinterface.AuthService {
	return &authService{
		repo: repo, users: users, sms: sms, tokens: tokens,
		cfg: cfg, roles: phoneRoles, log: log, now: time.Now,
	}
}

func (s *authService) RequestOTP(ctx context.Context, rawPhone string) (*serviceinterface.OTPRequestResult, error) {
	phone, err := validation.NormalizePhone(rawPhone)
	if err != nil {
		return nil, exception.NewInvalidPhone()
	}

	code, err := security.GenerateNumericCode(s.cfg.Length)
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}
	salt, err := security.NewSalt()
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}

	now := s.now()
	otp := &model.OTP{
		Phone:     phone,
		Salt:      salt,
		CodeHash:  security.HashSecret(salt, code),
		CreatedAt: now,
		ExpiresAt: now.Add(s.cfg.TTL),
	}
	blockedBy, err := s.repo.Save(ctx, otp, s.cfg.ResendCooldown)
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}
	if blockedBy != nil {
		retryAfter := blockedBy.CreatedAt.Add(s.cfg.ResendCooldown).Sub(now)
		logger.FromContext(ctx, s.log).Info("otp resend blocked by cooldown",
			"phone", validation.MaskPhone(phone), "retry_after_s", exception.RetryAfterSeconds(retryAfter))
		return nil, exception.NewOTPResendTooSoon(retryAfter)
	}

	body := fmt.Sprintf("Your Too Good To Go verification code is %s. It expires in %d minutes.",
		code, int(s.cfg.TTL.Minutes()))
	if err := s.sms.Send(ctx, phone, body); err != nil {
		// The code never reached the user, so do not leave it valid.
		_ = s.repo.Delete(ctx, phone)
		return nil, exception.NewInternalServer(fmt.Errorf("send sms: %w", err))
	}

	logger.FromContext(ctx, s.log).Info("otp requested", "phone", validation.MaskPhone(phone))
	res := &serviceinterface.OTPRequestResult{Phone: phone, TTL: s.cfg.TTL}
	if s.cfg.ExposeCode {
		res.DebugCode = code
	}
	return res, nil
}

func (s *authService) VerifyOTP(ctx context.Context, rawPhone, code string) (*serviceinterface.AuthResult, error) {
	phone, err := validation.NormalizePhone(rawPhone)
	if err != nil {
		return nil, exception.NewInvalidPhone()
	}
	// A malformed code can never match, so reject it without using up one of
	// the user's attempts (e.g. a typo with a missing digit).
	code = strings.TrimSpace(code)
	if !validation.IsNumericCode(code, s.cfg.Length) {
		return nil, exception.NewOTPMalformed(s.cfg.Length)
	}
	log := logger.FromContext(ctx, s.log).With("phone", validation.MaskPhone(phone))

	// Count the attempt before comparing so parallel guesses cannot exceed the limit.
	otp, err := s.repo.RecordAttempt(ctx, phone)
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}
	if otp == nil || otp.Used || otp.Expired(s.now()) {
		return nil, exception.NewOTPExpired()
	}
	if otp.Attempts > s.cfg.MaxAttempts {
		log.Warn("otp attempts exceeded")
		return nil, exception.NewOTPAttemptsExceeded()
	}

	if !security.SecretMatches(otp.Salt, code, otp.CodeHash) {
		log.Info("otp verification failed", "attempts", otp.Attempts)
		return nil, exception.NewOTPInvalid(s.cfg.MaxAttempts - otp.Attempts)
	}

	ok, err := s.repo.MarkUsed(ctx, phone)
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}
	if !ok { // a parallel request consumed it first
		return nil, exception.NewOTPExpired()
	}

	user, err := s.findOrCreateUser(ctx, phone)
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}
	accessToken, expiresIn, err := s.tokens.Issue(user)
	if err != nil {
		return nil, exception.NewInternalServer(err)
	}

	log.Info("otp verified", "user_id", user.ID, "role", string(user.Role))
	return &serviceinterface.AuthResult{User: user, AccessToken: accessToken, ExpiresIn: expiresIn}, nil
}

func (s *authService) findOrCreateUser(ctx context.Context, phone string) (*model.User, error) {
	role, ok := s.roles[phone]
	if !ok {
		role = enum.RoleCustomer
	}
	id, err := security.NewID()
	if err != nil {
		return nil, err
	}
	return s.users.FindOrCreateByPhone(ctx, &model.User{
		ID:        id,
		Phone:     phone,
		Role:      role,
		CreatedAt: s.now().UTC(),
	})
}
