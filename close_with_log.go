package app

import (
	"context"
	"io"
	"log/slog"
	"time"
)

// CloseWithLog closes the given io.Closer and logs any error that occurs to slog.
//
//	if err := closeable.Close(); err != nil {
//	   slog.Error("Error closing resource", "serviceName", serviceName, "err", err)
//	}
func CloseWithLog(closeable io.Closer, serviceName string) {
	if err := closeable.Close(); err != nil {
		slog.Error("Error closing resource", "serviceName", serviceName, "err", err)
	}
}

func RetryableCloseWithLog(closeable io.Closer, serviceName string) {
	maxRetries := 5
	retryDelay := time.Second
	startTime := time.Now()

	for i := 0; i < maxRetries; i++ {
		err := closeable.Close()
		if err == nil {
			return
		}

		slog.Error("Error closing resource, potential leak. Retrying...", "serviceName", serviceName, "err", err, "attempt", i+1, "elapsedTime", time.Since(startTime))
		time.Sleep(retryDelay)
		retryDelay *= 2
	}
}

func CloseWithLogWithContextDeadline(ctx context.Context, closeable io.Closer, serviceName string) {
	doneCh := make(chan struct{})
	go func() {
		if err := closeable.Close(); err != nil {
			slog.Error("Error closing resource", "serviceName", serviceName, "err", err)
		}
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-ctx.Done():
		slog.Warn("Closing resource timed out or canceled", "serviceName", serviceName, "err", ctx.Err())
	}
}
