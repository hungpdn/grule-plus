package logger

import (
	"context"
	"errors"
	"sync"
)

// Level represents the logging level.
type Level int

// Define the logging levels.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Instance represents the type of logger instance to use.
type Instance int

// Define the logger instances.
const (
	SlogInstance Instance = iota
	ZerologInstance
)

var (
	errInvalidLoggerInstance = errors.New("invalid logger instance")
)

// Attrs represents a map of attributes for structured logging.
// It is used to pass additional context or metadata with log messages.
type Attrs map[string]any

// Ilogger is an interface that defines the methods for logging at different levels.
// It allows for structured logging with attributes.
type ILogger interface {
	// Debugf logs a message at Debug level.
	Debugf(format string, args ...any)
	// Infof logs a message at Info level.
	Infof(format string, args ...any)
	// Warnf logs a message at Warn level.
	Warnf(format string, args ...any)
	// Errorf logs a message at Error level.
	Errorf(format string, args ...any)
	// WithAttrs returns a new logger with the given attributes added.
	WithAttrs(fields Attrs) ILogger
}

// logger is the global slog.Logger instance.
// use a sync.Once to ensure thread-safe initialization.
var (
	log  ILogger
	once sync.Once
)

// Config holds the configuration for the logger.
type Config struct {
	Level      Level
	JSONFormat bool
}

// init initializes the global logger instance.
func init() {
	once.Do(func() {
		err := NewLogger(Config{Level: LevelDebug, JSONFormat: false}, SlogInstance)
		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}
	})
}

// NewLogger initializes the logger with the given configuration and instance type.
// It supports different logger instances like slog and zerolog.
// Returns an error if the instance type is invalid.
func NewLogger(config Config, instance Instance) error {
	switch instance {
	case SlogInstance:
		log = newSlogLogger(config)
		return nil
	default:
		return errInvalidLoggerInstance
	}
}

// Debugf logs a message at Debug level with the given format and arguments.
func Debugf(format string, args ...any) {
	log.Debugf(format, args...)
}

// Infof logs a message at Info level with the given format and arguments.
func Infof(format string, args ...any) {
	log.Infof(format, args...)
}

// Warnf logs a message at Warn level with the given format and arguments.
func Warnf(format string, args ...any) {
	log.Warnf(format, args...)
}

// Errorf logs a message at Error level with the given format and arguments.
func Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

// WithAttrs returns a new logger with the given attributes added.
func WithAttrs(attrs Attrs) ILogger {
	return log.WithAttrs(attrs)
}

// WithContext returns a logger that includes context-specific attributes.
func WithContext(ctx context.Context) ILogger {
	if ctx != nil {
		attrs := Attrs{}
		if correlationId, ok := ctx.Value(CorrelationIdCtxKey).(string); ok {
			attrs[CorrelationIdCtxKey.String()] = correlationId
		}
		return log.WithAttrs(attrs)
	}
	return log
}
