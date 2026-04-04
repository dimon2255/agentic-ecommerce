package circuitbreaker

import (
	"testing"
	"time"
)

func TestCircuitBreaker_StartsClosedAndAllows(t *testing.T) {
	cb := New(3, 30*time.Second, 60*time.Second)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if cb.GetState() != Closed {
		t.Fatal("expected Closed state")
	}
}

func TestCircuitBreaker_TripsAfterThreshold(t *testing.T) {
	cb := New(3, 30*time.Second, 60*time.Second)

	// 3 failures should trip the breaker
	for range 3 {
		cb.RecordFailure()
	}

	if cb.GetState() != Open {
		t.Fatal("expected Open state after 3 failures")
	}
	if err := cb.Allow(); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestCircuitBreaker_SuccessResetsClosed(t *testing.T) {
	cb := New(3, 30*time.Second, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	// After success, counter should be reset
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != Closed {
		t.Fatal("expected Closed — success should have reset counter")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := New(3, 30*time.Second, 1*time.Millisecond) // very short open duration

	for range 3 {
		cb.RecordFailure()
	}

	// Wait for open duration to expire
	time.Sleep(5 * time.Millisecond)

	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil (half-open probe), got %v", err)
	}
	if cb.GetState() != HalfOpen {
		t.Fatal("expected HalfOpen state")
	}
}

func TestCircuitBreaker_HalfOpenSuccessCloses(t *testing.T) {
	cb := New(3, 30*time.Second, 1*time.Millisecond)

	for range 3 {
		cb.RecordFailure()
	}
	time.Sleep(5 * time.Millisecond)
	cb.Allow() // transition to half-open

	cb.RecordSuccess()
	if cb.GetState() != Closed {
		t.Fatal("expected Closed after half-open success")
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := New(3, 30*time.Second, 1*time.Millisecond)

	for range 3 {
		cb.RecordFailure()
	}
	time.Sleep(5 * time.Millisecond)
	cb.Allow() // transition to half-open

	cb.RecordFailure()
	if cb.GetState() != Open {
		t.Fatal("expected Open after half-open failure")
	}
}
