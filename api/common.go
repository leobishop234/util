package api

import (
	"encoding/json"
	"net/http"

	"github.com/leobishop234/util/observe"
	"github.com/leobishop234/util/srverr"
)

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		observe.SetLoggingMetadata(r.Context(), "response-err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	serErr, ok := srverr.Unwrap(err)
	if !ok {
		observe.SetLoggingMetadata(r.Context(), observe.LogMetadataErrKey, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	observe.SetLoggingMetadata(r.Context(), observe.LogMetadataErrKey, err.Error())
	WriteJSON(w, r, ErrCodeToHTTPStatus(serErr.Code()), serErr)
}

func ErrCodeToHTTPStatus(code srverr.ErrCode) int {
	switch code {
	case srverr.ErrCodeValidation:
		return http.StatusBadRequest
	case srverr.ErrCodeAuthenticationFailure:
		return http.StatusUnauthorized
	case srverr.ErrCodeAuthorizationFailure:
		return http.StatusForbidden
	case srverr.ErrCodeTimeout:
		return http.StatusRequestTimeout
	case srverr.ErrCodePreconditionFailed:
		return http.StatusPreconditionFailed
	case srverr.ErrCodeStateConflict:
		return http.StatusConflict
	case srverr.ErrCodePurged:
		return http.StatusGone
	case srverr.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case srverr.ErrCodeDependencyFailure:
		return http.StatusServiceUnavailable
	case srverr.ErrCodeNotFound:
		return http.StatusNotFound
	case srverr.ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
