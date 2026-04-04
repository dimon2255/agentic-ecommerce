package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

// AssistantRateLimiter enforces per-user and per-guest rate limits for assistant routes.
// It uses a sliding window approach with dual time horizons (per-minute burst + per-hour sustained).
type AssistantRateLimiter struct {
	mu               sync.Mutex
	windows          map[string]*slidingWindow
	userPerHour      int
	userPerMinute    int
	guestPerHour     int
	guestPerMinute   int
	lastCleanup      time.Time
	cleanupInterval  time.Duration
}

type slidingWindow struct {
	timestamps []time.Time
}

// NewAssistantRateLimiter creates a rate limiter for assistant routes.
func NewAssistantRateLimiter(userPerHour, userPerMinute, guestPerHour, guestPerMinute int) *AssistantRateLimiter {
	return &AssistantRateLimiter{
		windows:         make(map[string]*slidingWindow),
		userPerHour:     userPerHour,
		userPerMinute:   userPerMinute,
		guestPerHour:    guestPerHour,
		guestPerMinute:  guestPerMinute,
		lastCleanup:     time.Now(),
		cleanupInterval: 10 * time.Minute,
	}
}

// Middleware returns an HTTP middleware that enforces assistant-specific rate limits.
// It must be applied AFTER OptionalAuth so that user ID is available in context.
func (rl *AssistantRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, hasUser := GetUserID(r.Context())
		sessionID := r.Header.Get("X-Session-ID")

		var key string
		var hourLimit, minuteLimit int

		if hasUser {
			key = "user:" + userID
			hourLimit = rl.userPerHour
			minuteLimit = rl.userPerMinute
		} else if sessionID != "" {
			key = "guest:" + sessionID
			hourLimit = rl.guestPerHour
			minuteLimit = rl.guestPerMinute
		} else {
			response.Error(w, http.StatusUnauthorized, "authentication or session ID required")
			return
		}

		now := time.Now()

		rl.mu.Lock()

		// Lazy cleanup of stale windows
		if now.Sub(rl.lastCleanup) >= rl.cleanupInterval {
			cutoff := now.Add(-1 * time.Hour)
			for k, win := range rl.windows {
				if len(win.timestamps) == 0 || win.timestamps[len(win.timestamps)-1].Before(cutoff) {
					delete(rl.windows, k)
				}
			}
			rl.lastCleanup = now
		}

		win, exists := rl.windows[key]
		if !exists {
			win = &slidingWindow{}
			rl.windows[key] = win
		}

		// Prune timestamps older than 1 hour
		hourAgo := now.Add(-1 * time.Hour)
		pruned := win.timestamps[:0]
		for _, ts := range win.timestamps {
			if ts.After(hourAgo) {
				pruned = append(pruned, ts)
			}
		}
		win.timestamps = pruned

		// Count requests in last minute
		minuteAgo := now.Add(-1 * time.Minute)
		minuteCount := 0
		for i := len(win.timestamps) - 1; i >= 0; i-- {
			if win.timestamps[i].After(minuteAgo) {
				minuteCount++
			} else {
				break
			}
		}

		hourCount := len(win.timestamps)

		// Check minute burst limit first (more restrictive short-term)
		if minuteCount >= minuteLimit {
			retryAfter := win.timestamps[len(win.timestamps)-minuteCount].Add(time.Minute).Sub(now)
			rl.mu.Unlock()
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded — too many requests per minute")
			return
		}

		// Check hourly limit
		if hourCount >= hourLimit {
			retryAfter := win.timestamps[0].Add(time.Hour).Sub(now)
			rl.mu.Unlock()
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			response.Error(w, http.StatusTooManyRequests, "rate limit exceeded — hourly limit reached")
			return
		}

		// Allow request — record timestamp
		win.timestamps = append(win.timestamps, now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
