package middleware

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/leobishop234/util/log"
	"github.com/rs/zerolog"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
	body        bytes.Buffer
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	if r.status >= http.StatusBadRequest {
		r.body.Write(b)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func Logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := newStatusRecorder(w)

			metadata := LoggingMetadata{}
			ctx := context.WithValue(r.Context(), LoggingMetadataKey{}, metadata)
			r = r.WithContext(ctx)

			startedAt := time.Now()

			next.ServeHTTP(recorder, r)

			logEvent, message := eventForStatus(logger, recorder.status)
			logEvent = logEvent.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", recorder.status).
				Int("bytes", recorder.bytes).
				Dur("duration", time.Since(startedAt)).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent())

			if recorder.status >= http.StatusBadRequest {
				if body := strings.TrimSpace(recorder.body.String()); body != "" {
					logEvent = logEvent.Err(errors.New(body))
				}
			}

			if metadata := GetLoggingMetadata(r.Context()); len(metadata) > 0 { //nolint:contextcheck // false positive, context is read only here
				for key, value := range metadata {
					logEvent = logEvent.Any(key, value)
				}
			}

			logEvent.Msg(message)
		})
	}
}

func eventForStatus(logger *log.Logger, status int) (logEvent *zerolog.Event, message string) {
	if status < 300 {
		return logger.Info(), "http request successful"
	}
	if status < 500 {
		return logger.Warn(), "http request failed"
	}
	return logger.Error(), "http request errored"
}

type LoggingMetadataKey struct{}
type LoggingMetadata map[string]any

func GetLoggingMetadata(ctx context.Context) LoggingMetadata {
	return ctx.Value(LoggingMetadataKey{}).(LoggingMetadata)
}

var (
	LogErrKey   = "error"
	PanicErrKey = "panic"
)

func SetLoggingMetadata(ctx context.Context, key string, value any) {
	GetLoggingMetadata(ctx)[key] = value
}
