package utils

import (
	"context"
	"runtime/debug"

	"github.com/hungpdn/grule-plus/internal/logger"
)

// RecoverWithContext recovers from a panic and logs the stack trace with the provided context.
func RecoverWithContext(ctx context.Context, name string) {
	if r := recover(); r != nil {
		logger.WithContext(ctx).Errorf("%v panic : %v", name, string(debug.Stack()))
	}
}
