package logger

import (
	"io"
	"log/slog"
	"strings"
)

// NewLogger creates a new logger instance with the specified application name, output writer, log level, and version.
// It uses a text handler for development and a JSON handler for production environments.
// Includes source information (filename and line numbers) in log output.
func NewLogger(applicationName string,
	output io.Writer,
	level slog.Level,
	version string,
	appEnv string,
) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // This adds filename and line number to logs
	}
	var handler slog.Handler // handler is of type interface
	handler = slog.NewTextHandler(output, opts)
	if strings.ToLower(appEnv) == "production" {
		handler = slog.NewJSONHandler(output, opts)
	}

	logger := slog.New(handler)
	logger = logger.With(slog.String("application", applicationName)).
		With(slog.String("version", version))
	return logger
}
