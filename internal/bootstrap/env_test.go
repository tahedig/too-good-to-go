package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Chdir(t.TempDir()) // no .env file here
	for _, k := range []string{"APP_ENV", "HTTP_PORT", "LOG_LEVEL", "HTTP_READ_TIMEOUT", "TRUSTED_PROXIES"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppEnv != EnvDevelopment || cfg.HTTPPort != 8080 || cfg.LogLevel != "info" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.HTTPReadTimeout != 10*time.Second {
		t.Fatalf("unexpected read timeout: %v", cfg.HTTPReadTimeout)
	}
}

func TestLoadConfigReportsAllInvalidValues(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("APP_ENV", "prod")
	t.Setenv("HTTP_PORT", "99999")
	t.Setenv("HTTP_READ_TIMEOUT", "ten seconds")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, key := range []string{"APP_ENV", "HTTP_PORT", "HTTP_READ_TIMEOUT"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error should mention %s, got: %v", key, err)
		}
	}
}

func TestDotEnvDoesNotOverrideRealEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	content := "HTTP_PORT=9000\nLOG_LEVEL=\"debug\" \n# comment\nexport APP_ENV=staging # inline\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HTTP_PORT", "7000") // real env wins
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("APP_ENV")
	t.Cleanup(func() { os.Unsetenv("LOG_LEVEL"); os.Unsetenv("APP_ENV") })

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != 7000 {
		t.Errorf("real env should win, got port %d", cfg.HTTPPort)
	}
	if cfg.LogLevel != "debug" || cfg.AppEnv != EnvStaging {
		t.Errorf(".env values not applied: %+v", cfg)
	}
}
