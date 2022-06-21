package logger

import (
	"context"

	"github.com/no-mole/neptune/logger/level"

	"github.com/no-mole/neptune/logger/entry"

	"github.com/no-mole/neptune/logger/dispatcher"
)

type Logger interface {
	Fatal(ctx context.Context, tag string, err error, fields ...entry.Field)

	Error(ctx context.Context, tag string, err error, fields ...entry.Field)

	Warning(ctx context.Context, tag string, err error, fields ...entry.Field)

	Info(ctx context.Context, tag string, fields ...entry.Field)

	Notice(ctx context.Context, tag string, fields ...entry.Field)

	Debug(ctx context.Context, tag string, fields ...entry.Field)

	Trace(ctx context.Context, tag string, fields ...entry.Field)

	SetLevel(level.Level)

	Flush()

	AddDispatcher(dispatchers ...dispatcher.Dispatcher)

	AddHandle(ctx context.Context, handle Handle)

	Stop()
}

type Handle func(ctx context.Context, e entry.Entry) []entry.Field

func WithField(key string, value interface{}) entry.Field {
	return func(e entry.Entry) {
		e.WithField(key, value)
	}
}
