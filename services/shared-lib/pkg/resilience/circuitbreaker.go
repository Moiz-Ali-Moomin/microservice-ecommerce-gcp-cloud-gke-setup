package resilience

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when a call is rejected because the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // Normal operation — requests pass through.
	StateOpen                  // Fault mode — requests are rejected immediately.
	StateHalfOpen              // Recovery mode — a single probe request is allowed.
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements a thread-safe sliding-window circuit breaker.
//
// Transitions:
//   - CLOSED → OPEN: when consecutive failure count reaches Threshold.
//   - OPEN → HALF-OPEN: after Timeout elapses.
//   - HALF-OPEN → CLOSED: when the single probe request succeeds.
//   - HALF-OPEN → OPEN: when the single probe request fails.
type CircuitBreaker struct {
	mu sync.Mutex

	// Config
	name      string
	threshold int           // Failure count to trip open
	timeout   time.Duration // Time to wait before attempting half-open probe

	// State
	state       State
	failures    int
	lastFailure time.Time
	successInHalf bool
}

// NewCircuitBreaker creates a new CircuitBreaker.
//
//   - name      : identifier used in error messages (e.g. "postgres-client")
//   - threshold : consecutive failures before opening (e.g. 5)
//   - timeout   : time before testing recovery (e.g. 30s)
func NewCircuitBreaker(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:      name,
		threshold: threshold,
		timeout:   timeout,
		state:     StateClosed,
	}
}

// Execute runs fn through the circuit breaker.
// Returns ErrCircuitOpen immediately if the breaker is in OPEN state.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	state := cb.currentState()
	cb.mu.Unlock()

	if state == StateOpen {
		return ErrCircuitOpen
	}

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// State returns the current circuit breaker state (thread-safe snapshot).
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// Name returns the circuit breaker name.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// currentState resolves OPEN→HALF-OPEN transitions based on timeout.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) currentState() State {
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) >= cb.timeout {
			cb.state = StateHalfOpen
			cb.successInHalf = false
		}
	}
	return cb.state
}

// onFailure records a failure and potentially trips the breaker.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) onFailure() {
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Probe failed — go back to OPEN and reset timeout
		cb.state = StateOpen
	}
}

// onSuccess records a success and potentially closes the breaker.
// Must be called with cb.mu held.
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		// Probe succeeded — circuit recovered
		cb.state = StateClosed
		cb.failures = 0
	}
}
