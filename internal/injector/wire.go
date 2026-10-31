//go:build wireinject

package injector

import (
	"log/slog"

	"github.com/google/wire"

	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/handler"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/middleware"
	"github.com/tahedig/too-good-to-go/backend/internal/presentation/router"
)

func InitializeApp(cfg *bootstrap.Config, log *slog.Logger, version string) (*App, error) {
	wire.Build(
		provideHealthService,
		provideTokenManager,
		handler.NewHealthHandler,
		handler.NewProtectedHandler,
		middleware.NewErrorHandler,
		middleware.NewAuthMiddleware,
		provideHandlers,
		router.New,
		newApp,
	)
	return nil, nil
}
