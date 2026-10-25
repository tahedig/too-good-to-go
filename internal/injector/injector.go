// Package injector wires the dependency graph by hand. When the graph grows
// (repositories, services per feature) this can be replaced by google/wire
// (wire.go + generated wire_gen.go) without touching other layers.
package injector

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/tahedig/too-good-to-go/backend/internal/application/service"
	serviceinterface "github.com/tahedig/too-good-to-go/backend/internal/application/service/interface"
	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/handler"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/middleware"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/router"
)

type App struct {
	Router *gin.Engine
	Health serviceinterface.HealthService
}

func InitializeApp(cfg *bootstrap.Config, log *slog.Logger, version string) (*App, error) {
	// Dependency checks (Postgres, Redis, ...) are registered here once those
	// infrastructure packages exist, e.g. NewHealthService(..., pgChecker, redisChecker).
	healthService := service.NewHealthService(version, cfg.HealthCheckTimeout, log)

	r, err := router.New(cfg, log, middleware.NewErrorHandler(log), router.Handlers{
		Health: handler.NewHealthHandler(healthService),
	})
	if err != nil {
		return nil, err
	}

	return &App{Router: r, Health: healthService}, nil
}
