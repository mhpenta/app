package app

import (
	"log/slog"
	"time"
)

// LogSince logs the time since the start time, to be used ergonomically with defer.
func LogSince(msg string, start time.Time) {
	slog.Info(msg, "time", time.Since(start))
}
