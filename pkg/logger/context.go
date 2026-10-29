package logger

import (
	"context"
	"log/slog"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userIDKey
)

// WithRequestID stores the request id in ctx so code in any layer can log it
// without depending on gin.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithUserID is meant for the JWT middleware once authentication exists.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

// FromContext returns base enriched with request_id and user_id when present.
// Usage in a service: logger.FromContext(ctx, s.log).Info("offer created", "offer_id", id)
func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if base == nil {
		base = slog.Default()
	}
	if id := RequestID(ctx); id != "" {
		base = base.With("request_id", id)
	}
	if id := UserID(ctx); id != "" {
		base = base.With("user_id", id)
	}
	return base
}
