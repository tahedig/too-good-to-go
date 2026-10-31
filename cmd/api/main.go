// Package main is the API entrypoint.
//
// The comments below are the general Swagger info; run `make swagger` after
// changing them or any handler annotation.
//
//	@title						Too Good To Go API
//	@version					0.1
//	@description				Backend API. Every error uses the same body: {"error": {"code", "message", "details", "request_id"}}. Clients should switch on error.code.
//	@BasePath					/
//	@schemes					http https
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter "Bearer " followed by the access_token from /auth/otp/verify, e.g. "Bearer eyJhbGci..."
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
	"github.com/tahedig/too-good-to-go/backend/internal/injector"
	"github.com/tahedig/too-good-to-go/backend/pkg/logger"
)

// version is overridden at build time:
// go build -ldflags "-X main.version=v0.1.0" ./cmd/api
var version = "dev"

func main() {
	if err := run(); err != nil {
		// Even startup failures are logged as JSON.
		logger.New("error", os.Stderr).Error("server exited with error", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := bootstrap.LoadConfig()
	if err != nil {
		return err
	}

	log := logger.New(cfg.LogLevel, os.Stdout).With(
		"service", bootstrap.AppName,
		"env", cfg.AppEnv,
		"version", version,
	)
	slog.SetDefault(log)

	app, err := injector.InitializeApp(cfg, log, version)
	if err != nil {
		return err
	}
	// Runs after srv.Shutdown below, once no request can use the connections.
	defer func() {
		if err := app.Close(); err != nil {
			log.Warn("closing dependencies failed", "error", err.Error())
		}
	}()

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           app.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	log.Info("shutdown signal received, draining connections", "timeout", cfg.HTTPShutdownTimeout.String())
	app.Health.MarkShuttingDown()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Info("server stopped")
	return nil
}
