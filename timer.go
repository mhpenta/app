package app

import (
	"log/slog"
	"time"
)

// LogSince logs the elapsed time since a given start time. It's designed to be used with
// defer to easily measure and log function execution duration.
//
// Example usage:
//
//	func MyFunction() {
//	    defer LogSince("MyFunction completed in", time.Now())
//	    // ... function body ...
//	}
//
// The timing measurement will be logged when the function returns, showing the total
// execution time.
func LogSince(msg string, start time.Time) {
	slog.Info(msg, "time", time.Since(start))
}
