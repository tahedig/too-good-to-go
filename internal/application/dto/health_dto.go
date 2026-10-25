package dto

import "time"

const (
	HealthStatusOK           = "ok"
	HealthStatusUnavailable  = "unavailable"
	HealthStatusShuttingDown = "shutting_down"
)

type HealthResponse struct {
	Status    string                       `json:"status"`
	Version   string                       `json:"version"`
	Timestamp time.Time                    `json:"timestamp"`
	Checks    map[string]HealthCheckResult `json:"checks,omitempty"`
}

// HealthCheckResult deliberately has no error text: /health is public, so
// failure reasons are logged server side instead of exposed.
type HealthCheckResult struct {
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}
