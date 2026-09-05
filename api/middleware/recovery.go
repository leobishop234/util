package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/leobishop234/util/observe"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func(context context.Context) {
				if recovered := recover(); recovered != nil {
					observe.SetLoggingMetadata(context, observe.LogMetadataPanicKey, fmt.Sprint(recovered))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}(r.Context())

			next.ServeHTTP(w, r)
		})
	}
}
