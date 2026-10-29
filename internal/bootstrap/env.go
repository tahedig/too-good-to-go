package bootstrap

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string

	HTTPHost            string
	HTTPPort            int
	HTTPReadTimeout     time.Duration
	HTTPWriteTimeout    time.Duration
	HTTPIdleTimeout     time.Duration
	HTTPShutdownTimeout time.Duration

	HealthCheckTimeout time.Duration

	LogLevel string

	// TrustedProxies is the list of proxy IPs/CIDRs allowed to set
	// X-Forwarded-For. Empty means no proxy is trusted.
	TrustedProxies []string
}

// Addr returns the host:port the HTTP server listens on.
func (c *Config) Addr() string {
	return net.JoinHostPort(c.HTTPHost, strconv.Itoa(c.HTTPPort))
}

func (c *Config) IsProduction() bool { return c.AppEnv == EnvProduction }

// LoadConfig reads configuration from the environment. If a .env file exists in
// the working directory it is loaded first, but real environment variables
// always win over values in the file.
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	p := &envParser{}
	cfg := &Config{
		AppEnv:              p.oneOf("APP_ENV", EnvDevelopment, EnvDevelopment, EnvStaging, EnvProduction),
		HTTPHost:            p.string("HTTP_HOST", ""),
		HTTPPort:            p.port("HTTP_PORT", 8080),
		HTTPReadTimeout:     p.duration("HTTP_READ_TIMEOUT", 10*time.Second),
		HTTPWriteTimeout:    p.duration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		HTTPIdleTimeout:     p.duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		HTTPShutdownTimeout: p.duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		HealthCheckTimeout:  p.duration("HEALTH_CHECK_TIMEOUT", 2*time.Second),
		LogLevel:            p.oneOf("LOG_LEVEL", "info", "debug", "info", "warn", "error"),
		TrustedProxies:      p.list("TRUSTED_PROXIES"),
	}

	if err := errors.Join(p.errs...); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

// envParser collects every problem so the developer sees all invalid
// variables at once instead of fixing them one by one.
type envParser struct {
	errs []error
}

func (p *envParser) string(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(v)
	}
	return def
}

func (p *envParser) oneOf(key, def string, allowed ...string) string {
	v := strings.ToLower(p.string(key, def))
	if v == "" {
		return def
	}
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	p.errs = append(p.errs, fmt.Errorf("%s=%q must be one of %v", key, v, allowed))
	return def
}

func (p *envParser) port(key string, def int) int {
	raw := p.string(key, "")
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 65535 {
		p.errs = append(p.errs, fmt.Errorf("%s=%q must be a port number between 1 and 65535", key, raw))
		return def
	}
	return n
}

func (p *envParser) duration(key string, def time.Duration) time.Duration {
	raw := p.string(key, "")
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		p.errs = append(p.errs, fmt.Errorf("%s=%q must be a positive duration such as 5s or 1m", key, raw))
		return def
	}
	return d
}

func (p *envParser) list(key string) []string {
	raw := p.string(key, "")
	if raw == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}
