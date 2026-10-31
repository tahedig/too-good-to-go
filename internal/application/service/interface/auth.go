package serviceinterface

import (
	"context"
	"time"

	"github.com/tahedig/too-good-to-go/backend/internal/domain/model"
)

// OTPRequestResult is returned after a code was issued and sent.
type OTPRequestResult struct {
	Phone string // normalized, E.164
	TTL   time.Duration
	// DebugCode is the plain code, set only when OTP_EXPOSE_CODE is enabled.
	DebugCode string
}

// AuthResult is returned after a successful OTP verification.
type AuthResult struct {
	User        *model.User
	AccessToken string
	ExpiresIn   time.Duration
}

type AuthService interface {
	// RequestOTP issues a new OTP for phone (replacing any previous one) and
	// sends it by SMS.
	RequestOTP(ctx context.Context, phone string) (*OTPRequestResult, error)
	// VerifyOTP checks code against the active OTP and consumes it on success.
	// It then creates or loads the user and issues an access token.
	VerifyOTP(ctx context.Context, phone, code string) (*AuthResult, error)
}
