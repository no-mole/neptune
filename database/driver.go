package database

import (
	"context"
	"fmt"
	"time"

	"github.com/no-mole/neptune/logger"

	"gorm.io/gorm"
	grrmLogger "gorm.io/gorm/logger"
)

type Driver interface {
	Dial(conf *Config) gorm.Dialector
}

var driverMapping = map[string]Driver{}

func RegistryDriver(name string, driver Driver) {
	driverMapping[name] = driver
}

func GetDriver(name string) (_ Driver, exist bool) {
	if driver, ok := driverMapping[name]; ok {
		return driver, true
	}
	return nil, false
}

type Logger struct{}

func (l *Logger) LogMode(level grrmLogger.LogLevel) grrmLogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, s string, i ...interface{}) {
	logger.Info(ctx, "database", logger.WithField("msg", fmt.Sprintf(s, i...)))
}

func (l *Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	logger.Warning(ctx, "database", nil, logger.WithField("msg", fmt.Sprintf(s, i...)))
}

func (l *Logger) Error(ctx context.Context, s string, i ...interface{}) {
	logger.Error(ctx, "database", nil, logger.WithField("msg", fmt.Sprintf(s, i...)))
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
		logger.WithField("sql", sql),
		logger.WithField("rows", rows),
		logger.WithField("msg", errMsg),
	)
}

var _ grrmLogger.Interface = &Logger{}
