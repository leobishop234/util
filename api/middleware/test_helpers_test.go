package middleware

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func firstLogLine(t *testing.T, b *bytes.Buffer) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatal("expected at least one log line")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("failed to parse log line as JSON: %v", err)
	}

	return payload
}
