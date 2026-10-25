// Package logger builds the structured JSON logger used across the service.
package logger

import (
	"io"
	"log/slog"
	"strings"
)

const redacted = "[REDACTED]"

// sensitiveKeyParts: any attribute whose key contains one of these is masked,
// e.g. "password", "refresh_token", "Authorization", "otp_code".
var sensitiveKeyParts = []string{
	"password", "passwd", "secret", "token", "authorization",
	"cookie", "otp", "api_key", "apikey", "card_number", "cvv",
}

// New returns a JSON logger writing to w at the given level
// ("debug", "info", "warn", "error"). Unknown levels fall back to info.
func New(level string, w io.Writer) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       ParseLevel(level),
		ReplaceAttr: maskSensitive,
	})
	return slog.New(h)
}

func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func maskSensitive(_ []string, a slog.Attr) slog.Attr {
	if isSensitiveKey(a.Key) {
		return slog.String(a.Key, redacted)
	}
	return a
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	for _, part := range sensitiveKeyParts {
		if strings.Contains(k, part) {
			return true
		}
	}
	return false
}
