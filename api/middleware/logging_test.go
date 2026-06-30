package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestLoggingMiddleware_LogsResponseBodyOnError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantErrLog string
	}{
		{
			name:       "4xx logs response body as error",
			status:     http.StatusBadRequest,
			body:       "invalid input\n",
			wantErrLog: "invalid input",
		},
		{
			name:       "5xx logs response body as error",
			status:     http.StatusInternalServerError,
			body:       "Internal Server Error\n",
			wantErrLog: "Internal Server Error",
		},
		{
			name:   "2xx does not log response body",
			status: http.StatusOK,
			body:   "all good\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := zerolog.New(&output)

			middleware := Logging(&logger)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, strings.TrimSpace(tt.body), tt.status)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			logLine := firstLogLine(t, &output)

			if tt.wantErrLog != "" {
				if got, ok := logLine["error"].(string); !ok || got != tt.wantErrLog {
					t.Fatalf("expected error field %q, got %v", tt.wantErrLog, logLine["error"])
				}
			} else {
				if _, present := logLine["error"]; present {
					t.Fatalf("expected no error field, but got %v", logLine["error"])
				}
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
