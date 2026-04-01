package middleware

import "sync"

// WebhookReplayGuard tracks processed webhook event IDs to prevent replay attacks.
// Uses a bounded in-memory set — suitable for single-instance deployments.
type WebhookReplayGuard struct {
	mu       sync.Mutex
	seen     map[string]struct{}
	maxSize  int
}

// NewWebhookReplayGuard creates a guard with the given max tracked events.
func NewWebhookReplayGuard(maxSize int) *WebhookReplayGuard {
	return &WebhookReplayGuard{
		seen:    make(map[string]struct{}),
		maxSize: maxSize,
	}
}

// Check returns true if the event ID has already been processed.
// If not, it records the ID and returns false.
func (g *WebhookReplayGuard) Check(eventID string) bool {
	if eventID == "" {
		return false
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.seen[eventID]; exists {
		return true // already processed
	}

	// Evict oldest if at capacity (simple clear strategy)
	if len(g.seen) >= g.maxSize {
		g.seen = make(map[string]struct{})
	}

	g.seen[eventID] = struct{}{}
	return false
}
