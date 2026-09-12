package util

import (
	"log/slog"
	"os"
)

// NewLogger builds the application structured logger.
func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
