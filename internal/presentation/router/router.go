package router

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
	"github.com/tahedig/too-good-to-go/backend/internal/domain/enum"
	handlerinterface "github.com/tahedig/too-good-to-go/backend/internal/presentation/handler/interface"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/middleware"
	middlewareinterface "github.com/tahedig/too-good-to-go/backend/internal/presentation/middleware/interface"
)

// Handlers groups every HTTP handler. Add new feature handlers here.
type Handlers struct {
	Health    handlerinterface.HealthHandler
	Protected handlerinterface.ProtectedHandler
}

func New(
	cfg *bootstrap.Config,
	log *slog.Logger,
	errHandler middlewareinterface.ErrorHandler,
	authMW middlewareinterface.AuthMiddleware,
	h Handlers,
) (*gin.Engine, error) {
	// Release mode always: gin's debug output is plain text and would break
	// the JSON-only log stream.
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.HandleMethodNotAllowed = true
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("invalid TRUSTED_PROXIES: %w", err)
	}

	// Order matters: the logger sits outside Recover so panics are logged
	// with their final 500 status.
	r.Use(
		middleware.RequestID(),
		middleware.RequestLogger(log),
		errHandler.Recover(),
		errHandler.Handle(),
	)
	r.NoRoute(errHandler.NotFound())
	r.NoMethod(errHandler.MethodNotAllowed())

	r.GET(bootstrap.HealthPath, h.Health.Health)
	r.HEAD(bootstrap.HealthPath, h.Health.Health)

	r.GET("/auth/me", authMW.Authenticate(), h.Protected.Me)

	// Test routes, one per role, to exercise RBAC.
	test := r.Group("/test", authMW.Authenticate())
	test.GET("/customer", authMW.RequireRoles(enum.RoleCustomer), h.Protected.CustomerArea)
	test.GET("/vendor", authMW.RequireRoles(enum.RoleVendor), h.Protected.VendorArea)
	test.GET("/admin", authMW.RequireRoles(enum.RoleAdmin), h.Protected.AdminArea)

	return r, nil
}
