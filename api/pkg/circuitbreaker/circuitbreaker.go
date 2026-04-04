package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit breaker is in the open state.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the circuit breaker state.
type State int

const (
	Closed   State = iota
	Open
	HalfOpen
)

// CircuitBreaker protects against cascading failures from external service outages.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	lastFailure  time.Time
	threshold    int
	window       time.Duration
	openDuration time.Duration
	openedAt     time.Time
}

// New creates a circuit breaker.
// threshold: consecutive failures within window to trip.
// window: time window for counting failures.
// openDuration: how long to stay open before allowing a probe.
func New(threshold int, window, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        Closed,
		threshold:    threshold,
		window:       window,
		openDuration: openDuration,
	}
}

// Allow returns true if the request should proceed.
// Returns false (and ErrOpen) when the circuit is open.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Closed:
		return nil
	case Open:
		if time.Since(cb.openedAt) >= cb.openDuration {
			cb.state = HalfOpen
			return nil
		}
		return ErrOpen
	case HalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess records a successful call. Transitions half-open to closed.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.state == HalfOpen {
		cb.state = Closed
	}
}

// RecordFailure records a failed call. May transition to open state.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Reset failure count if last failure was outside the window
	if now.Sub(cb.lastFailure) > cb.window {
		cb.failures = 0
	}

	cb.failures++
	cb.lastFailure = now

	if cb.failures >= cb.threshold {
		cb.state = Open
		cb.openedAt = now
	}

	// Half-open probe failed — go back to open
	if cb.state == HalfOpen {
		cb.state = Open
		cb.openedAt = now
	}
}

// State returns the current circuit breaker state.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
