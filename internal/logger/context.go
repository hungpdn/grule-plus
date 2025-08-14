package logger

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

type ContextKey string

func (ck ContextKey) String() string {
	return string(ck)
}

// CorrelationIdCtxKey is the context key for storing correlation IDs.
const (
	CorrelationIdCtxKey = ContextKey("correlation_id")
)

// GetCorrelationIdFromCtx retrieves the correlation ID from the context.
func GetCorrelationIdFromCtx(ctx context.Context) string {
	return GetStringFromCtx(ctx, CorrelationIdCtxKey)
}

// SetCorrelationIdToCtx sets a new correlation ID in the context if it does not already exist.
func SetCorrelationIdToCtx(ctx context.Context) context.Context {
	if GetCorrelationIdFromCtx(ctx) != "" {
		return ctx
	}
	return context.WithValue(ctx, CorrelationIdCtxKey, NewCorrelationID())
}

// GetStringFromCtx retrieves a string value from the context using the specified key.
func GetStringFromCtx(ctx context.Context, key ContextKey) string {
	if ctx != nil {
		if val, ok := ctx.Value(key).(string); ok {
			return val
		}
	}
	return ""
}

// NewCorrelationID generates a new correlation ID using ULID.
func NewCorrelationID() string {
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
