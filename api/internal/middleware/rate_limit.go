package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// RateLimiter implements a simple token bucket rate limiter per IP.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int           // tokens per interval
	interval time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a rate limiter with the given rate per interval.
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		interval: interval,
	}
}

// Middleware returns an HTTP middleware that enforces the rate limit.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		rl.mu.Lock()
		b, exists := rl.buckets[ip]
		if !exists {
			b = &bucket{tokens: rl.rate, lastReset: time.Now()}
			rl.buckets[ip] = b
		}

		// Reset tokens if interval has passed
		if time.Since(b.lastReset) >= rl.interval {
			b.tokens = rl.rate
			b.lastReset = time.Now()
		}

		if b.tokens <= 0 {
			retryAfter := rl.interval - time.Since(b.lastReset)
			rl.mu.Unlock()

			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		b.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (client IP)
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}
