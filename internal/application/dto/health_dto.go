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

type HealthCheckResult struct {
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}
