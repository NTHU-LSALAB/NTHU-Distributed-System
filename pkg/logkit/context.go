package logkit

import (
	"context"
	"log"
)

// Inject logger into context to easily carry the logger through everywhere

type loggerConextKey int8

const contextKeyLogger loggerConextKey = iota

func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(contextKeyLogger).(*Logger)
	if !ok {
		log.Fatal("logger is not found in context")
	}

	return logger
}
