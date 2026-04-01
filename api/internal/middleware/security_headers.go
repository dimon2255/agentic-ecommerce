package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders sets standard security headers on every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS only for non-localhost
		host := r.Host
		if !strings.HasPrefix(host, "localhost") && !strings.HasPrefix(host, "127.0.0.1") {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// CSP: allow Stripe JS, Google Fonts, Unsplash images
		w.Header().Set("Content-Security-Policy", strings.Join([]string{
			"default-src 'self'",
			"script-src 'self' https://js.stripe.com",
			"frame-src https://js.stripe.com",
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
			"font-src https://fonts.gstatic.com",
			"img-src 'self' https://images.unsplash.com data:",
			"connect-src 'self' https://api.stripe.com",
		}, "; "))

		next.ServeHTTP(w, r)
	})
}
