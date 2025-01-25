package retry

import (
	"context"
	"github.com/mhpenta/app"
	"math/rand/v2"
	"time"
)

// Config retry mechanism
type Config struct {
	// Number of retries to be applied
	Times int
	// InitialDelayMillisecond is the delay before the first retry
	InitialDelayMilliseconds int
	// ExponentialBackoff function that calculates the retry delay
	ExponentialBackoff func(retryCount int) time.Duration
}

func NewConfig(retryCount int) Config {
	return Config{
		Times:              retryCount,
		ExponentialBackoff: ExponentialBackoff1sPower2,
	}
}

// Execute the task and retries when the task returns an error
func Execute[T any](ctx context.Context, config Config, task func(ctx context.Context) (T, error)) (T, error) {
	var mRetryErr app.MultiError
	var defaultResult T

	for i := 0; i < config.Times; i++ {
		result, err := task(ctx)

		if err == nil {
			return result, nil
		} else {
			mRetryErr.Errors = append(mRetryErr.Errors, err)
		}

		if i == config.Times-1 {
			break
		}

		var delay time.Duration

		if config.ExponentialBackoff != nil {
			delay = config.ExponentialBackoff(i + 1)
		} else {
			delay = ExponentialBackoff1sPower2(i + 1)
		}

		select {
		case <-ctx.Done():
			return defaultResult, mRetryErr.ErrorOrNil()
		case <-time.After(delay * time.Millisecond):
		}
	}

	return defaultResult, mRetryErr.ErrorOrNil()
}

// ExecuteWithTwoReturns the task and retries when the task returns an error
func ExecuteWithTwoReturns[T1, T2 any](ctx context.Context, config Config, task func(ctx context.Context) (T1, T2, error)) (T1, T2, error) {
	var mRetryErr app.MultiError
	var defaultResult1 T1
	var defaultResult2 T2

	for i := 0; i < config.Times; i++ {
		result1, result2, err := task(ctx)

		if err == nil {
			return result1, result2, nil
		} else {
			mRetryErr.Errors = append(mRetryErr.Errors, err)
		}

		if i == config.Times-1 {
			break
		}

		var delay time.Duration

		if config.ExponentialBackoff != nil {
			delay = config.ExponentialBackoff(i + 1)
		} else {
			delay = ExponentialBackoff1sPower2(i + 1)
		}

		select {
		case <-ctx.Done():
			return defaultResult1, defaultResult2, mRetryErr.ErrorOrNil()
		case <-time.After(delay * time.Millisecond):
		}
	}

	return defaultResult1, defaultResult2, mRetryErr.ErrorOrNil()
}

// ExponentialBackoff1sPower2 calculates the delay as an exponential backoff of 1 second, power of 2
func ExponentialBackoff1sPower2(retryCount int) time.Duration {
	// Start with a 100ms delay and double it with each retry
	return time.Duration(1000*(1<<retryCount)) * time.Millisecond
}

// ExponentialBackoff3sPower2 calculates the delay as an exponential backoff of 1 second, power of 2
func ExponentialBackoff3sPower2(retryCount int) time.Duration {
	// Start with a 100ms delay and double it with each retry
	return time.Duration(3000*(1<<retryCount)) * time.Millisecond
}

// ExponentialBackoff1sPower2WithJitter calculates the delay as an exponential backoff of 1 second, power of 2, with jitter
func ExponentialBackoff1sPower2WithJitter(retryCount int) time.Duration {
	baseDelay := time.Duration(1000*(1<<retryCount)) * time.Millisecond
	jitter := 0.5
	maxDelay := baseDelay + time.Duration(float64(baseDelay)*jitter)
	delay := baseDelay + time.Duration(rand.N(int64(maxDelay-baseDelay)))
	return delay
}
