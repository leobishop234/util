package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leobishop234/util/api/middleware"
	"github.com/leobishop234/util/srverr"
	"github.com/stretchr/testify/require"
)

func requestWithLoggingMetadata(method, target string) (*http.Request, middleware.LoggingMetadata) {
	metadata := middleware.LoggingMetadata{}
	req := httptest.NewRequest(method, target, nil)
	ctx := context.WithValue(req.Context(), middleware.LoggingMetadataKey{}, metadata)
	return req.WithContext(ctx), metadata
}

func TestWriteJSONSetsJSONContentType(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req, metadata := requestWithLoggingMetadata(http.MethodGet, "/users")
	WriteJSON(recorder, req, http.StatusCreated, map[string]string{"name": "alice"})

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "alice", response["name"])
	require.Empty(t, metadata)
}

func TestWriteErrorSetsJSONContentType(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req, metadata := requestWithLoggingMetadata(http.MethodPost, "/bookings")
	WriteError(recorder, req, srverr.New(srverr.ErrCodeValidation, "invalid request", nil))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Empty(t, response)
	require.Equal(t, "code: 2, message: invalid request", metadata[middleware.LogErrKey])
}

func TestWriteErrorSetsMetadataForGenericError(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("something went wrong")
	recorder := httptest.NewRecorder()
	req, metadata := requestWithLoggingMetadata(http.MethodGet, "/status")
	WriteError(recorder, req, genericErr)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, "something went wrong\n", recorder.Body.String())
	require.Equal(t, "something went wrong", metadata[middleware.LogErrKey])
}

func TestWriteJSONSetsResponseErrorMetadataWhenEncodingFails(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	req, metadata := requestWithLoggingMetadata(http.MethodGet, "/users")
	WriteJSON(recorder, req, http.StatusCreated, map[string]any{
		"invalid": func() {},
	})

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	responseErr, ok := metadata["response-err"].(string)
	require.True(t, ok)
	require.Contains(t, responseErr, "unsupported type")
}

func TestErrCodeToHTTPStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		code       srverr.ErrCode
		wantStatus int
	}{
		{
			name:       "maps internal to 500",
			code:       srverr.ErrCodeInternal,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "maps validation to 400",
			code:       srverr.ErrCodeValidation,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "maps authentication failure to 401",
			code:       srverr.ErrCodeAuthenticationFailure,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "maps authorisation failure to 403",
			code:       srverr.ErrCodeAuthorizationFailure,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "maps timeout to 408",
			code:       srverr.ErrCodeTimeout,
			wantStatus: http.StatusRequestTimeout,
		},
		{
			name:       "maps precondition failed to 412",
			code:       srverr.ErrCodePreconditionFailed,
			wantStatus: http.StatusPreconditionFailed,
		},
		{
			name:       "maps state conflict to 409",
			code:       srverr.ErrCodeStateConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "maps purged to 410",
			code:       srverr.ErrCodePurged,
			wantStatus: http.StatusGone,
		},
		{
			name:       "maps rate limited to 429",
			code:       srverr.ErrCodeRateLimited,
			wantStatus: http.StatusTooManyRequests,
		},
		{
			name:       "maps dependency failure to 503",
			code:       srverr.ErrCodeDependencyFailure,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "maps not found to 404",
			code:       srverr.ErrCodeNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "maps unknown code to 500",
			code:       srverr.ErrCode(999),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.wantStatus, ErrCodeToHTTPStatus(tc.code))
		})
	}
}
