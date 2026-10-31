package router_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tahedig/too-good-to-go/backend/internal/application/service"
	"github.com/tahedig/too-good-to-go/backend/internal/bootstrap"
)

func TestSwaggerServedWhenEnabled(t *testing.T) {
	e := newAuthEnvWith(t, &bootstrap.Config{SwaggerEnabled: true},
		service.OTPConfig{Length: 6, TTL: time.Minute, MaxAttempts: 3})

	if w := e.get("/swagger", ""); w.Code != http.StatusMovedPermanently ||
		w.Header().Get("Location") != "/swagger/index.html" {
		t.Fatalf("/swagger should redirect to the UI: %d %q", w.Code, w.Header().Get("Location"))
	}

	w := e.get("/swagger/index.html", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "swagger-ui") {
		t.Fatalf("index.html: %d", w.Code)
	}
	if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "script-src 'self'") {
		t.Fatalf("swagger needs the relaxed CSP, got %q", csp)
	}

	w = e.get("/swagger/doc.json", "")
	for _, path := range []string{"/auth/otp/request", "/auth/otp/verify", "/auth/me", "BearerAuth"} {
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), path) {
			t.Fatalf("doc.json should describe %s: %d", path, w.Code)
		}
	}

	// API routes keep the strict policy.
	if csp := e.get("/health", "").Header().Get("Content-Security-Policy"); !strings.HasPrefix(csp, "default-src 'none'") {
		t.Fatalf("API CSP was relaxed: %q", csp)
	}
}

func TestSwaggerHiddenWhenDisabled(t *testing.T) {
	e := newAuthEnv(t, time.Minute, 3) // SwaggerEnabled is false
	if w := e.get("/swagger/index.html", ""); w.Code != http.StatusNotFound {
		t.Fatalf("swagger must be off: %d", w.Code)
	}
}
