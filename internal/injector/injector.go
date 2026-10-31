// Package injector wires the dependency graph by hand. When the graph grows
// (repositories, services per feature) this can be replaced by google/wire
// (wire.go + generated wire_gen.go) without touching other layers.
package injector

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tahedig/too-good-to-go/backend/internal/application/service"
	serviceinterface "github.com/tahedig/too-good-to-go/backend/internal/application/service/interface"
	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/enum"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/repository"
	"github.com/tahedig/too-good-to-go/backend/internal/infrastructure/repository/memory"
	redisrepo "github.com/tahedig/too-good-to-go/backend/internal/infrastructure/repository/redis"
	"github.com/tahedig/too-good-to-go/backend/internal/infrastructure/sms"
	"github.com/tahedig/too-good-to-go/backend/internal/infrastructure/token"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/handler"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/middleware"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/router"
)

type App struct {
	Router *gin.Engine
	Health serviceinterface.HealthService
	// Close releases external connections; call it after the HTTP server stopped.
	Close func() error
}

// storage is everything that moves from memory to Redis when REDIS_URL is set.
type storage struct {
	otps     repository.OTPRepository
	limiter  serviceinterface.RateLimiter
	checkers []serviceinterface.HealthChecker
	close    func() error
}

const redisConnectTimeout = 5 * time.Second

func newStorage(cfg *bootstrap.Config, log *slog.Logger) (*storage, error) {
	if cfg.RedisURL == "" {
		// bootstrap only allows this in development.
		log.Warn("REDIS_URL is not set: OTPs and rate limits are kept in memory, are lost on restart and are not shared between instances")
		return &storage{
			otps:    memory.NewOTPRepository(),
			limiter: memory.NewRateLimiter(),
			close:   func() error { return nil },
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), redisConnectTimeout)
	defer cancel()
	client, err := redisrepo.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	log.Info("connected to redis", "addr", client.Options().Addr, "db", client.Options().DB)
	return &storage{
		otps:     redisrepo.NewOTPRepository(client, cfg.RedisKeyPrefix),
		limiter:  redisrepo.NewRateLimiter(client, cfg.RedisKeyPrefix),
		checkers: []serviceinterface.HealthChecker{redisrepo.NewHealthChecker(client)},
		close:    client.Close,
	}, nil
}

func newSMSSender(cfg *bootstrap.Config, log *slog.Logger) serviceinterface.SMSSender {
	switch cfg.SMSProvider {
	// Add real providers here (e.g. case bootstrap.SMSProviderFerzz). The
	// value is validated in bootstrap, so the default is unreachable.
	default:
		return sms.NewFakeSender(log, !cfg.IsProduction())
	}
}

func InitializeApp(cfg *bootstrap.Config, log *slog.Logger, version string) (*App, error) {
	store, err := newStorage(cfg, log)
	if err != nil {
		return nil, err
	}

	healthService := service.NewHealthService(version, cfg.HealthCheckTimeout, log, store.checkers...)
	smsSender := newSMSSender(cfg, log)
	tokens := token.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessTTL)

	phoneRoles := make(map[string]enum.Role)
	for _, p := range cfg.VendorPhones {
		phoneRoles[p] = enum.RoleVendor
	}
	for _, p := range cfg.AdminPhones { // admin wins if a number is in both lists
		phoneRoles[p] = enum.RoleAdmin
	}

	authService := service.NewAuthService(
		// Users stay in memory until the Postgres repository exists.
		store.otps, memory.NewUserRepository(), smsSender, tokens,
		service.OTPConfig{
			Length:         cfg.OTPLength,
			TTL:            cfg.OTPTTL,
			MaxAttempts:    cfg.OTPMaxAttempts,
			ResendCooldown: cfg.OTPResendCooldown,
			ExposeCode:     cfg.OTPExposeCode,
		},
		phoneRoles, log)

	r, err := router.New(cfg, log, middleware.NewErrorHandler(log), middleware.NewAuthMiddleware(tokens), store.limiter, router.Handlers{
		Health:    handler.NewHealthHandler(healthService),
		Auth:      handler.NewAuthHandler(authService),
		Protected: handler.NewProtectedHandler(),
	})
	if err != nil {
		_ = store.close()
		return nil, err
	}

	return &App{Router: r, Health: healthService, Close: store.close}, nil
}
