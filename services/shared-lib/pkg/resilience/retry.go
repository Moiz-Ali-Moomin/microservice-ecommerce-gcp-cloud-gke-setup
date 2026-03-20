package resilience

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// ErrMaxRetries is returned when all retry attempts have been exhausted.
var ErrMaxRetries = errors.New("max retries exceeded")

// RetryConfig holds the parameters for exponential backoff retry.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (1 = no retry).
	MaxAttempts int
	// InitialDelay is the wait time before the first retry.
	InitialDelay time.Duration
	// MaxDelay caps the backoff duration.
	MaxDelay time.Duration
	// Multiplier is the factor by which the delay grows each retry (e.g. 2.0 = doubles).
	Multiplier float64
	// Jitter adds a random fraction of the delay to prevent thundering herd (0.0–1.0).
	Jitter float64
	// RetryIf is called to determine if a specific error warrants a retry.
	// If nil, all errors are retried.
	RetryIf func(err error) bool
}

// DefaultRetryConfig is a sensible production default: 3 attempts, 100ms initial, 5s max.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:  3,
	InitialDelay: 100 * time.Millisecond,
	MaxDelay:     5 * time.Second,
	Multiplier:   2.0,
	Jitter:       0.2,
}

// Do executes fn with exponential backoff retry according to cfg.
// It respects context cancellation between attempts.
//
// Example:
//
//	err := resilience.Do(ctx, resilience.DefaultRetryConfig, func(ctx context.Context) error {
//	    return callExternalService(ctx)
//	})
func Do(ctx context.Context, cfg RetryConfig, fn func(ctx context.Context) error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}

	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		// If RetryIf is set, only retry when it returns true
		if cfg.RetryIf != nil && !cfg.RetryIf(lastErr) {
			return lastErr
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		// Wait with exponential backoff + jitter before next attempt
		jitterAmount := time.Duration(float64(delay) * cfg.Jitter * rand.Float64()) //nolint:gosec
		sleepDuration := delay + jitterAmount
		if sleepDuration > cfg.MaxDelay {
			sleepDuration = cfg.MaxDelay
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDuration):
		}

		// Grow the delay for the next attempt
		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return errors.Join(ErrMaxRetries, lastErr)
}
