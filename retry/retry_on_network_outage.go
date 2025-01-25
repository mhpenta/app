package retry

import (
	"context"
	"fmt"
	"log/slog"
	"modeledge-go/ext/httpext"
	"time"
)

// NetworkRetryConfig holds configuration for the retry mechanism
type NetworkRetryConfig struct {
	MaxAttempts int
	SleepTime   time.Duration
	MaxWaitTime time.Duration
}

// DefaultNetworkRetryConfig provides sensible default values for RetryConfig
var DefaultNetworkRetryConfig = NetworkRetryConfig{
	MaxAttempts: 480,
	SleepTime:   1 * time.Minute,
	MaxWaitTime: 8 * time.Hour,
}

// OnNetworkError retries the given function with a standard wait time on network errors with default configuration
//
// Function is designed to re-attempt a function if the error it encounters is a network error, typically due to a
// network outage.
//
// See retry.DefaultNetworkRetryConfig for defaults.
func OnNetworkError[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	return OnNetworkErrorWithConfig(ctx, f, DefaultNetworkRetryConfig)
}

// OnNetworkErrorWithConfig retries the given function with a standard wait time on network errors
func OnNetworkErrorWithConfig[T any](ctx context.Context, f func(context.Context) (T, error), config NetworkRetryConfig) (T, error) {
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

			if !httpext.IsDialError(err) {
				return result, err
			}

			attempt++
			if attempt >= config.MaxAttempts {
				return result, fmt.Errorf("max retry attempts reached: %w", err)
			}

			if time.Since(startTime) > config.MaxWaitTime {
				return result, fmt.Errorf("max wait time exceeded: %w", err)
			}

			slog.Info("Network unreachable, retrying",
				"error", err,
				"attempt", attempt,
				"nextRetryIn", waitDuration,
			)
			time.Sleep(waitDuration)
		}
	}
}

// OnNetworkErrorOnlyError retries the given function with a standard wait time on network errors with default configuration
//
// Function is designed to re-attempt a function if the error it encounters is a network error, typically due to a
// network outage.
//
// See retry.DefaultNetworkRetryConfig for defaults.
func OnNetworkErrorOnlyError(ctx context.Context, f func(context.Context) error) error {
	return OnNetworkErrorWithConfigOnlyError(ctx, f, DefaultNetworkRetryConfig)
}

// OnNetworkErrorWithConfigOnlyError retries the given function with a standard wait time on network errors
func OnNetworkErrorWithConfigOnlyError(ctx context.Context, f func(context.Context) error, config NetworkRetryConfig) error {

	var err error

	startTime := time.Now()
	attempt := 0
	waitDuration := config.SleepTime

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, aborting retry", "error", ctx.Err())
			return ctx.Err()
		default:
			err = f(ctx)
			if err == nil {
				return nil
			}

			if !httpext.IsDialError(err) {
				return err
			}

			attempt++
			if attempt >= config.MaxAttempts {
				return fmt.Errorf("max retry attempts reached: %w", err)
			}

			if time.Since(startTime) > config.MaxWaitTime {
				return fmt.Errorf("max wait time exceeded: %w", err)
			}

			slog.Info("Network unreachable, retrying",
				"error", err,
				"attempt", attempt,
				"nextRetryIn", waitDuration,
			)
			time.Sleep(waitDuration)
		}
	}
}
