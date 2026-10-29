package serviceinterface

import (
	"context"

	"github.com/tahedig/too-good-to-go/backend/internal/application/dto"
)

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

type HealthService interface {
	Check(ctx context.Context) dto.HealthResponse

	MarkShuttingDown()
}
