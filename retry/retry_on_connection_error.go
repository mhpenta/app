package retry

import (
	"context"
	"fmt"
	"github.com/mhpenta/app/httpext"
	"log/slog"
	"time"
)

// ConnectionRetryConfig holds configuration for the retry mechanism
type ConnectionRetryConfig struct {
	MaxAttempts int
	SleepTime   time.Duration
	MaxWaitTime time.Duration
}

// DefaultConnectionRetryConfig provides sensible default values for RetryConfig
var DefaultConnectionRetryConfig = ConnectionRetryConfig{
	MaxAttempts: 20,
	SleepTime:   30 * time.Second,
	MaxWaitTime: 6 * time.Minute,
}

// OnConnectionError retries the given function with a standard wait time on Connection errors with default configuration
//
// Function is designed to re-attempt a function if the error it encounters is a Connection error.
//
// This is designed to be a simplistic solution when dealing with APIs.
//
// See retry.DefaultConnectionRetryConfig for defaults.
func OnConnectionError[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	return OnConnectionErrorWithConfig(ctx, f, DefaultConnectionRetryConfig)
}

// OnConnectionErrorWithConfig retries the given function with a standard wait time on Connection errors
func OnConnectionErrorWithConfig[T any](ctx context.Context, f func(context.Context) (T, error), config ConnectionRetryConfig) (T, error) {
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

			if !httpext.IsTransientNetworkOrDNSIssueErr(err) || !httpext.IsDialError(err) {
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

func OnConnectionErrorSimple(ctx context.Context, f func() error) error {
	return OnConnectionErrorSimpleWithConfig(ctx, f, DefaultConnectionRetryConfig)
}

// OnConnectionErrorSimpleWithConfig retries the given function with a standard wait time on Connection errors
func OnConnectionErrorSimpleWithConfig(ctx context.Context, f func() error, config ConnectionRetryConfig) error {
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
			err = f()
			if err == nil {
				return nil
			}

			if !httpext.IsTransientNetworkOrDNSIssueErr(err) || !httpext.IsDialError(err) {
				return err
			}

			attempt++
			if attempt >= config.MaxAttempts {
				return fmt.Errorf("max retry attempts reached: %w", err)
			}

			if time.Since(startTime) > config.MaxWaitTime {
				return fmt.Errorf("max wait time exceeded: %w", err)
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
