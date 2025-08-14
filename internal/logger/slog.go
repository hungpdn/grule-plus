package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// slogLogger is an implementation of ILogger using the slog package.
type slogLogger struct {
	log *slog.Logger
}

// getSlogLevel converts the custom Level type to slog.Level.
func getSlogLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New creates a new slog.Logger instance based on the provided configuration.
func newSlogLogger(cfg Config) *slogLogger {
	var handler slog.Handler

	if cfg.JSONFormat {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     getSlogLevel(cfg.Level),
			AddSource: true,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     getSlogLevel(cfg.Level),
			AddSource: true,
		})
	}

	return &slogLogger{slog.New(handler)}
}

// Debugf logs a message at Debug level.
func (l *slogLogger) Debugf(format string, args ...any) {
	l.log.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a message at Info level.
func (l *slogLogger) Infof(format string, args ...any) {
	l.log.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a message at Warn level.
func (l *slogLogger) Warnf(format string, args ...any) {
	l.log.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a message at Error level.
func (l *slogLogger) Errorf(format string, args ...any) {
	l.log.Error(fmt.Sprintf(format, args...))
}

// WithAttrs returns a new logger with the given attributes added.
func (l *slogLogger) WithAttrs(fields Attrs) ILogger {
	var f = make([]any, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.log.With(f...)
	return &slogLogger{newLogger}
}
