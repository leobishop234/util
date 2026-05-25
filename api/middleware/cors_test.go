package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_AddsHeadersAndCallsNext(t *testing.T) {
	middleware := CORS()

	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/location/tags", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if handlerCalled {
		t.Fatal("expected wrapped handler not to be called for preflight requests")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected Access-Control-Allow-Origin to echo request origin, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "GET,POST,PUT,PATCH,DELETE,OPTIONS" {
		t.Fatalf("expected default Access-Control-Allow-Methods header, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type,Authorization" {
		t.Fatalf("expected default Access-Control-Allow-Headers header, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("expected Access-Control-Max-Age header, got %q", got)
	}
	if got := rr.Header().Values("Vary"); len(got) != 3 {
		t.Fatalf("expected 3 Vary header values, got %d", len(got))
	}
}
