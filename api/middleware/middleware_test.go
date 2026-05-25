package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestMiddlewares_OrderRecoversAndLogs(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)

	var handler http.Handler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	for _, middleware := range middlewares(&logger) {
		handler = middleware(handler)
	}

	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}

	logLine := firstLogLine(t, &output)
	if got, _ := logLine["message"].(string); got != "http request errored" {
		t.Fatalf("expected log message %q, got %q", "http request errored", got)
	}
	if got, _ := logLine["panic"].(string); got != "boom" {
		t.Fatalf("expected panic field %q, got %q", "boom", got)
	}
}
