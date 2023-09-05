package logger

import (
	"context"
	"go.uber.org/zap"
)

type Logger interface {
	Fatal(ctx context.Context, msg string, err error, fields ...zap.Field)

	Error(ctx context.Context, msg string, err error, fields ...zap.Field)

	Warning(ctx context.Context, msg string, err error, fields ...zap.Field)

	Info(ctx context.Context, msg string, fields ...zap.Field)

	Notice(ctx context.Context, msg string, fields ...zap.Field)

	Debug(ctx context.Context, msg string, fields ...zap.Field)

	Trace(ctx context.Context, msg string, fields ...zap.Field)

	SetLevel(Level)

	Handle(ctx context.Context, handle Handle)
}

type Handle func(ctx context.Context, lvl Level, msg string, err error, fields ...zap.Field) []zap.Field

func WithField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
