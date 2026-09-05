package middleware

import (
	"net/http"

	"github.com/leobishop234/util/observe"
)

// ApplyMiddleware wraps a handler with the default middleware stack.
func ApplyMiddleware(logger *observe.Logger, handler http.Handler) http.Handler {
	for _, mw := range Middlewares(logger) {
		handler = mw(handler)
	}

	return handler
}

// Middlewares returns the default middleware stack in execution order.
// This applies as: Context -> Logging -> Recovery -> Handler.
func Middlewares(logger *observe.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		CORS(),
		Recovery(),
		Logging(logger),
	}
}
