package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Moiz-Ali-Moomin/microservice-ecommerce-gcp-cloud-gke-setup/services/shared-lib/pkg/resilience"
)

var errRetriable = errors.New("transient error")

// ──────────────────────────────────────────────────────────────────────────────
// Retry tests
// ──────────────────────────────────────────────────────────────────────────────

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := resilience.Do(context.Background(), resilience.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
	}, func(ctx context.Context) error {
		calls++
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetry_SuccessAfterRetry(t *testing.T) {
	calls := 0
	err := resilience.Do(context.Background(), resilience.RetryConfig{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		Multiplier:   1,
	}, func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return errRetriable
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_ExhaustsMaxAttempts(t *testing.T) {
	calls := 0
	err := resilience.Do(context.Background(), resilience.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		Multiplier:   1,
	}, func(ctx context.Context) error {
		calls++
		return errRetriable
	})

	if !errors.Is(err, resilience.ErrMaxRetries) {
		t.Errorf("expected ErrMaxRetries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	err := resilience.Do(ctx, resilience.RetryConfig{
		MaxAttempts:  100,
		InitialDelay: 50 * time.Millisecond,
		Multiplier:   1,
	}, func(ctx context.Context) error {
		calls++
		return errRetriable
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetry_RetryIfPredicate(t *testing.T) {
	nonRetriableErr := errors.New("non-retriable")
	calls := 0

	err := resilience.Do(context.Background(), resilience.RetryConfig{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		RetryIf: func(err error) bool {
			return !errors.Is(err, nonRetriableErr)
		},
	}, func(ctx context.Context) error {
		calls++
		return nonRetriableErr
	})

	if !errors.Is(err, nonRetriableErr) {
		t.Errorf("expected nonRetriableErr, got %v", err)
	}
	// Should stop immediately without retrying
	if calls != 1 {
		t.Errorf("expected 1 call (no retry), got %d", calls)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Circuit breaker tests
// ──────────────────────────────────────────────────────────────────────────────

func TestCircuitBreaker_StartsInClosedState(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", 3, time.Second)
	if cb.CurrentState() != resilience.StateClosed {
		t.Errorf("expected CLOSED state, got %s", cb.CurrentState())
	}
}

func TestCircuitBreaker_TransitionsToOpen(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", 3, time.Second)

	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return errors.New("fail") }) //nolint:errcheck
	}

	if cb.CurrentState() != resilience.StateOpen {
		t.Errorf("expected OPEN state after %d failures, got %s", 3, cb.CurrentState())
	}
}

func TestCircuitBreaker_RejectsWhenOpen(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", 1, time.Hour)

	cb.Execute(func() error { return errors.New("fail") }) //nolint:errcheck

	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, resilience.ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", 1, 20*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") }) //nolint:errcheck
	time.Sleep(30 * time.Millisecond)

	if cb.CurrentState() != resilience.StateHalfOpen {
		t.Errorf("expected HALF-OPEN state, got %s", cb.CurrentState())
	}
}

func TestCircuitBreaker_RecoveryToClosedOnSuccess(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", 1, 10*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") }) //nolint:errcheck
	time.Sleep(20 * time.Millisecond)

	// Probe succeeds → should close
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected probe to succeed: %v", err)
	}

	if cb.CurrentState() != resilience.StateClosed {
		t.Errorf("expected CLOSED after recovery, got %s", cb.CurrentState())
	}
}

func TestCircuitBreaker_ProbeFailureReopens(t *testing.T) {
	cb := resilience.NewCircuitBreaker("test", 1, 10*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") }) //nolint:errcheck
	time.Sleep(20 * time.Millisecond)

	// Probe fails → should go back to OPEN
	cb.Execute(func() error { return errors.New("still broken") }) //nolint:errcheck

	if cb.CurrentState() != resilience.StateOpen {
		t.Errorf("expected OPEN after failed probe, got %s", cb.CurrentState())
	}
}
