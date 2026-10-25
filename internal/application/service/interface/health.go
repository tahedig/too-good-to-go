// Package serviceinterface holds the contracts the presentation layer depends
// on. ("interface" is a Go keyword, so the package name differs from the folder.)
package serviceinterface

import (
	"context"

	"github.com/tahedig/too-good-to-go/backend/internal/application/dto"
)

// HealthChecker is implemented by any dependency that should affect /health
// (Postgres, Redis, RabbitMQ...). Infrastructure packages provide them and the
// injector registers them; the health service never imports infrastructure.
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

type HealthService interface {
	Check(ctx context.Context) dto.HealthResponse
	// MarkShuttingDown makes /health report 503 so load balancers stop
	// routing traffic while in-flight requests finish.
	MarkShuttingDown()
}
