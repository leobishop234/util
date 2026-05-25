package middleware

import (
	"net/http"

	"github.com/leobishop234/util/log"
)

// ApplyMiddleware wraps a handler with the default middleware stack.
func ApplyMiddleware(logger *log.Logger, handler http.Handler) http.Handler {
	for _, mw := range middlewares(logger) {
		handler = mw(handler)
	}

	return handler
}

// middlewares returns the default middleware stack in execution order.
// This applies as: Context -> Logging -> Recovery -> Handler.
func middlewares(logger *log.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		CORS(),
		Recovery(),
		Logging(logger),
	}
}
