package tools

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker protects downstream services/tools from cascading failures.
// States:
// - StateClosed: Normal operations. Calls pass through.
// - StateOpen: Tripped due to failures. Calls fail fast immediately.
// - StateHalfOpen: Cooldown period expired. Lets a probe call test the waters.
type CircuitBreaker struct {
	mu              sync.Mutex
	state           State
	failures        int
	maxFailures     int
	resetTimeout    time.Duration
	lastFailureTime time.Time
}

// NewCircuitBreaker initializes a circuit breaker in the closed state.
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// State returns the current state of the circuit breaker (thread-safe).
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
} 


// Execute wraps an arbitrary tool call with circuit breaker safety.
func (cb *CircuitBreaker) Execute(fn func() (string, error)) (string, error) {
	cb.mu.Lock()

	// 1. If the circuit is open, check if the cooldown timeout has elapsed.
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			// Cooldown period over -> allow a single probe through to test if the service recovered
			cb.state = StateHalfOpen
		} else {
			// Still within cooldown -> fail fast without executing fn
			cb.mu.Unlock()
			return "", ErrCircuitOpen
		}
	}

	// 2. CRITICAL: Release the lock before running fn!
	// fn() could be a slow HTTP or LLM call (e.g. 2-5 seconds).
	// If we held the lock here, all other concurrent workers would freeze waiting for this lock.
	cb.mu.Unlock()

	res, err := fn()

	// 3. Re-acquire the lock to record results and safely update state.
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		// If probe failed during HalfOpen OR failures exceeded threshold in Closed -> trip open!
		if cb.failures >= cb.maxFailures || cb.state == StateHalfOpen {
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}
		return "", err
	}

	// On success: reset failure counter and restore Closed (healthy) state.
	cb.failures = 0
	cb.state = StateClosed

	return res, nil
}



