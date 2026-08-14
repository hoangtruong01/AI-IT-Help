package logger

import (
	"log/slog"
	"os"
)

// InitLogger configures and returns a structured JSON logger using standard log/slog.
func InitLogger(serviceName, env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
		slog.String("env", env),
	)

	slog.SetDefault(logger)
	return logger
}
