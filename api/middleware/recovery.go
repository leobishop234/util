package middleware

import (
	"context"
	"fmt"
	"net/http"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func(context context.Context) {
				if recovered := recover(); recovered != nil {
					SetLoggingMetadata(context, PanicErrKey, fmt.Sprint(recovered))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}(r.Context())

			next.ServeHTTP(w, r)
		})
	}
}
