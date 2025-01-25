package retry

import (
	"context"
	"fmt"
	"log/slog"
	"modeledge-go/ext/jsonext"
	"time"
)

// UnmarshallingRetryConfig holds configuration for the retry mechanism
type UnmarshallingRetryConfig struct {
	MaxAttempts int
	SleepTime   time.Duration
	MaxWaitTime time.Duration
}

// DefaultUnmarshallingErrorRetryConfig provides sensible default values for RetryConfig
var DefaultUnmarshallingErrorRetryConfig = UnmarshallingRetryConfig{
	MaxAttempts: 2,
	SleepTime:   7 * time.Second,
	MaxWaitTime: 30 * time.Minute,
}

// OnUnmarshallingError retries the given function with a standard wait time on Connection errors with default configuration
//
// Function is designed to re-attempt a function if the error it encounters is a Connection error.
//
// This is designed to be a simplistic solution when dealing with APIs.
//
// See retry.DefaultConnectionRetryConfig for defaults.
func OnUnmarshallingError[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	return OnUnmarshallingErrorWithConfig(ctx, f, DefaultUnmarshallingErrorRetryConfig)
}

// OnUnmarshallingErrorWithConfig retries the given function with a standard wait time on Connection errors
func OnUnmarshallingErrorWithConfig[T any](ctx context.Context, f func(context.Context) (T, error), config UnmarshallingRetryConfig) (T, error) {
	var result T
	var err error

	startTime := time.Now()
	attempt := 0
	waitDuration := config.SleepTime

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, aborting retry", "error", ctx.Err())
			return result, ctx.Err()
		default:
			result, err = f(ctx)
			if err == nil {
				return result, nil
			}

			if !jsonext.IsUnmarshallingError(err) {
				return result, err
			}

			attempt++
			if attempt >= config.MaxAttempts {
				return result, fmt.Errorf("max retry attempts reached: %w", err)
			}

			if time.Since(startTime) > config.MaxWaitTime {
				return result, fmt.Errorf("max wait time exceeded: %w", err)
			}

			slog.Info("Connection unreachable, retrying",
				"error", err,
				"attempt", attempt,
				"nextRetryIn", waitDuration,
			)
			time.Sleep(waitDuration)
		}
	}
}
