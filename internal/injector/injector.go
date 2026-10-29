// Package injector assembles the application dependency graph.
package injector

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/tahedig/too-good-to-go/backend/internal/application/service"
	serviceinterface "github.com/tahedig/too-good-to-go/backend/internal/application/service/interface"
	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
	handlerinterface "github.com/tahedig/too-good-to-go/backend/internal/presentation/handler/interface"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/router"
)

type App struct {
	Router *gin.Engine
	Health serviceinterface.HealthService
}

func provideHealthService(cfg *bootstrap.Config, log *slog.Logger, version string) serviceinterface.HealthService {
	return service.NewHealthService(version, cfg.HealthCheckTimeout, log)
}

func provideHandlers(health handlerinterface.HealthHandler) router.Handlers {
	return router.Handlers{Health: health}
}

func newApp(r *gin.Engine, health serviceinterface.HealthService) *App {
	return &App{Router: r, Health: health}
}
