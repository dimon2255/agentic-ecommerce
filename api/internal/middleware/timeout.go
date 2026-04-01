package middleware

import (
	"net/http"
	"time"
)

// Timeout wraps handlers with a request timeout. If the handler takes longer
// than the specified duration, a 503 Service Unavailable is returned.
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, duration, `{"error":"request timeout"}`)
	}
}
