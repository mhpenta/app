// Package app contains general application management functions
package app

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"syscall"
)

var ErrContextCancelled = errors.New("context has been cancelled or has expired")

// ContextCancelled is a utility function to check if a context has been cancelled.
func ContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		slog.Info("Context has been cancelled or has expired")
		return true
	default:
		return false
	}
}

// MainContext returns a context that is cancelled when the application receives an interrupt signal. It is the main
// application "background" context. It cancels on these signals: syscall.SIGINT, syscall.SIGKILL syscall.SIGTERM
func MainContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,  // os.Interrupt
		syscall.SIGKILL, // os.Kill
		syscall.SIGTERM)
}

func IsContextCancelledOrExpiredError(err error) bool {
	return errors.Is(err, ErrContextCancelled) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
