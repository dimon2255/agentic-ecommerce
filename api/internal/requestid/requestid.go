package requestid

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

type contextKey string

const key contextKey = "request_id"

const Header = "X-Request-ID"

// Middleware reads or generates a request ID, stores it in context, and sets the response header.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(Header)
		if id == "" {
			id = Generate()
		}
		w.Header().Set(Header, id)
		ctx := context.WithValue(r.Context(), key, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Get extracts the request ID from context. Returns "" if not set.
func Get(ctx context.Context) string {
	id, _ := ctx.Value(key).(string)
	return id
}

// Generate creates a new random ID in UUID-like format.
func Generate() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
