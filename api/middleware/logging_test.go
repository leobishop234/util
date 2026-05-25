package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggingMiddleware_LogLevelByStatusCode(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		wantLevel   string
		wantMessage string
	}{
		{
			name:        "2xx logs at info",
			status:      http.StatusOK,
			wantLevel:   "info",
			wantMessage: "http request successful",
		},
		{
			name:        "4xx logs at warn",
			status:      http.StatusBadRequest,
			wantLevel:   "warn",
			wantMessage: "http request failed",
		},
		{
			name:        "5xx logs at error",
			status:      http.StatusInternalServerError,
			wantLevel:   "error",
			wantMessage: "http request errored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := zerolog.New(&output)

			middleware := Logging(&logger)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))

			req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
			req.RemoteAddr = "127.0.0.1:1234"
			req.Header.Set("User-Agent", "middleware-test")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.status {
				t.Fatalf("expected status %d, got %d", tt.status, rr.Code)
			}

			logLine := firstLogLine(t, &output)
			if got, ok := logLine["level"].(string); !ok || got != tt.wantLevel {
				t.Fatalf("expected level %q, got %v", tt.wantLevel, logLine["level"])
			}
			if got, ok := logLine["message"].(string); !ok || got != tt.wantMessage {
				t.Fatalf("expected message %q, got %v", tt.wantMessage, logLine["message"])
			}
			if got, ok := logLine["method"].(string); !ok || got != http.MethodGet {
				t.Fatalf("expected method %q, got %v", http.MethodGet, logLine["method"])
			}
			if got, ok := logLine["path"].(string); !ok || got != "/healthz" {
				t.Fatalf("expected path %q, got %v", "/healthz", logLine["path"])
			}
			if got, ok := logLine["status"].(float64); !ok || int(got) != tt.status {
				t.Fatalf("expected status %d in log, got %v", tt.status, logLine["status"])
			}
		})
	}
}

func TestLoggingMiddleware_IncludesPanicFromContext(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)

	var handler http.Handler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	for _, middleware := range []func(http.Handler) http.Handler{
		Recovery(),
		Logging(&logger),
	} {
		handler = middleware(handler)
	}

	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	logLine := firstLogLine(t, &output)
	if got, ok := logLine["panic"].(string); !ok || got != "boom" {
		t.Fatalf("expected panic field %q, got %v", "boom", logLine["panic"])
	}
}
