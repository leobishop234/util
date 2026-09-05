package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/leobishop234/util/observe"
)

func TestRecoveryMiddleware_ReturnsInternalServerError(t *testing.T) {
	middleware := Recovery()
	handler := middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	req = req.WithContext(context.WithValue(
		req.Context(),
		observe.LoggingMetadataKey{},
		observe.LoggingMetadata{},
	))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), http.StatusText(http.StatusInternalServerError)) {
		t.Fatalf("expected body to contain %q, got %q", http.StatusText(http.StatusInternalServerError), rr.Body.String())
	}
}
