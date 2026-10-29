package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
)

// RequestLogger writes one structured JSON line per request. Query strings are
// not logged because they may carry tokens or OTP codes.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []slog.Attr{
			slog.String("request_id", c.GetString(bootstrap.CtxKeyRequestID)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", c.FullPath()),
			slog.Int("status", status),
			slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000),
			slog.Int("response_bytes", max(c.Writer.Size(), 0)),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}
		if uid := c.GetString(bootstrap.CtxKeyUserID); uid != "" {
			attrs = append(attrs, slog.String("user_id", uid))
		}

		level := slog.LevelInfo
		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case status >= http.StatusBadRequest:
			level = slog.LevelWarn
		case c.Request.URL.Path == bootstrap.HealthPath:
			level = slog.LevelDebug // probes hit this constantly
		}

		log.LogAttrs(c.Request.Context(), level, "http request", attrs...)
	}
}
