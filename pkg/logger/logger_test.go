package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestLoggerWritesJSONAndMasksSecrets(t *testing.T) {
	var buf bytes.Buffer
	log := New("info", &buf)

	ctx := WithUserID(WithRequestID(context.Background(), "req-1"), "user-42")
	FromContext(ctx, log).Info("login attempt", "phone", "0912", "password", "hunter2", "refresh_token", "abc")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not JSON: %v\n%s", err, buf.String())
	}
	want := map[string]any{
		"level": "INFO", "msg": "login attempt", "request_id": "req-1", "user_id": "user-42",
		"phone": "0912", "password": redacted, "refresh_token": redacted,
	}
	for k, v := range want {
		if entry[k] != v {
			t.Errorf("%s = %v, want %v", k, entry[k], v)
		}
	}
}

func TestDebugIsFilteredAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	New("info", &buf).Debug("noise")
	if buf.Len() != 0 {
		t.Fatalf("debug log should be dropped at info level, got %s", buf.String())
	}
}
