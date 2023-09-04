package database

import (
	"context"
	"fmt"
	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/logger/entry"
	grrmLogger "gorm.io/gorm/logger"
	"time"
)

func NewLogger(fields ...entry.Field) grrmLogger.Interface {
	return &Logger{Fields: fields}
}

type Logger struct {
	Fields []entry.Field
}

func (l *Logger) LogMode(level grrmLogger.LogLevel) grrmLogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, s string, i ...interface{}) {
	logger.Info(ctx, "database", append([]entry.Field{logger.WithField("msg", fmt.Sprintf(s, i...))}, l.Fields...)...)
}

func (l *Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	logger.Warning(ctx, "database", nil, append([]entry.Field{logger.WithField("msg", fmt.Sprintf(s, i...))}, l.Fields...)...)
}

func (l *Logger) Error(ctx context.Context, s string, i ...interface{}) {
	logger.Error(ctx, "database", nil, append([]entry.Field{logger.WithField("msg", fmt.Sprintf(s, i...))}, l.Fields...)...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	logger.Info(
		ctx,
		"database",
		append([]entry.Field{
			logger.WithField("sql", sql),
			logger.WithField("rows", rows),
			logger.WithField("msg", errMsg),
		}, l.Fields...)...,
	)
}

var _ grrmLogger.Interface = &Logger{}
