package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/tahedig/too-good-to-go/backend/internal/application/dto"
)

type fakeChecker struct {
	name string
	err  error
	wait time.Duration
}

func (f fakeChecker) Name() string { return f.name }
func (f fakeChecker) Check(ctx context.Context) error {
	if f.wait > 0 {
		select {
		case <-time.After(f.wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return f.err
}

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestHealthOKWithoutCheckers(t *testing.T) {
	resp := NewHealthService("v1", time.Second, discard).Check(context.Background())
	if resp.Status != dto.HealthStatusOK || resp.Version != "v1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHealthUnavailableWhenACheckFails(t *testing.T) {
	svc := NewHealthService("v1", time.Second, discard,
		fakeChecker{name: "postgres"},
		fakeChecker{name: "redis", err: errors.New("connection refused")},
	)
	resp := svc.Check(context.Background())
	if resp.Status != dto.HealthStatusUnavailable {
		t.Fatalf("status = %s, want unavailable", resp.Status)
	}
	if resp.Checks["postgres"].Status != dto.HealthStatusOK || resp.Checks["redis"].Status != dto.HealthStatusUnavailable {
		t.Fatalf("unexpected checks: %+v", resp.Checks)
	}
}

func TestHealthCheckTimeout(t *testing.T) {
	svc := NewHealthService("v1", 20*time.Millisecond, discard, fakeChecker{name: "slow", wait: time.Second})
	start := time.Now()
	resp := svc.Check(context.Background())
	if resp.Status != dto.HealthStatusUnavailable {
		t.Fatalf("slow dependency should be unavailable, got %s", resp.Status)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("timeout was not applied")
	}
}

func TestHealthShuttingDown(t *testing.T) {
	svc := NewHealthService("v1", time.Second, discard)
	svc.MarkShuttingDown()
	if got := svc.Check(context.Background()).Status; got != dto.HealthStatusShuttingDown {
		t.Fatalf("status = %s, want shutting_down", got)
	}
}
