package service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tahedig/too-good-to-go/backend/internal/application/dto"
	serviceinterface "github.com/tahedig/too-good-to-go/backend/internal/application/service/interface"
	"github.com/tahedig/too-good-to-go/backend/pkg/logger"
)

type healthService struct {
	version      string
	timeout      time.Duration
	checkers     []serviceinterface.HealthChecker
	log          *slog.Logger
	shuttingDown atomic.Bool
	now          func() time.Time
}

func NewHealthService(
	version string,
	timeout time.Duration,
	log *slog.Logger,
	checkers ...serviceinterface.HealthChecker,
) serviceinterface.HealthService {
	return &healthService{
		version:  version,
		timeout:  timeout,
		checkers: checkers,
		log:      log,
		now:      time.Now,
	}
}

func (s *healthService) MarkShuttingDown() { s.shuttingDown.Store(true) }

// Check runs all dependency checks concurrently, each with its own timeout.
// The overall status is "ok" only if every check passes.
func (s *healthService) Check(ctx context.Context) dto.HealthResponse {
	resp := dto.HealthResponse{
		Status:    dto.HealthStatusOK,
		Version:   s.version,
		Timestamp: s.now().UTC(),
	}

	if s.shuttingDown.Load() {
		resp.Status = dto.HealthStatusShuttingDown
		return resp
	}
	if len(s.checkers) == 0 {
		return resp
	}

	resp.Checks = make(map[string]dto.HealthCheckResult, len(s.checkers))
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, c := range s.checkers {
		wg.Add(1)
		go func(c serviceinterface.HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, s.timeout)
			defer cancel()

			start := time.Now()
			err := c.Check(checkCtx)
			result := dto.HealthCheckResult{
				Status:    dto.HealthStatusOK,
				LatencyMS: time.Since(start).Milliseconds(),
			}
			if err != nil {
				result.Status = dto.HealthStatusUnavailable
				logger.FromContext(ctx, s.log).Warn("health check failed",
					"check", c.Name(), "error", err.Error(), "latency_ms", result.LatencyMS)
			}

			mu.Lock()
			resp.Checks[c.Name()] = result
			if err != nil {
				resp.Status = dto.HealthStatusUnavailable
			}
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	return resp
}
