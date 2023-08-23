package database

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"sync"
	"time"

	validate "github.com/go-playground/validator/v10"
	"github.com/no-mole/neptune/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/opentelemetry/tracing"
)

var (
	//configMap map[string]*gorm.DB
	configMap sync.Map
	validator = validate.New()
)

type Config struct {
	Driver   string `json:"driver" yaml:"driver" validate:"required"`
	Host     string `json:"host" yaml:"host" validate:"required"`
	Database string `json:"database" yaml:"database" validate:"required"`
	Port     int    `json:"port" yaml:"port" validate:"required"`
	Username string `json:"username" yaml:"username" validate:"required"`
	Password string `json:"password" yaml:"password" validate:"required"`

	MaxIdleConnes int `json:"max_idle_connes" yaml:"max_idle_connes" validate:"required"`
	MaxOpenConnes int `json:"max_open_connes" yaml:"max_open_connes" validate:"required"`
	MaxLifetime   int `json:"max_lifetime" yaml:"max_lifetime" validate:"required"`

	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" validate:"required"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout" validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Struct(c)
}

type BaseDb struct {
	*gorm.DB
}

func (b *BaseDb) SetEngine(ctx context.Context, engine string) bool {
	if value, ok := configMap.Load(engine); ok {
		if db, ok := value.(*gorm.DB); ok {
			b.DB = db.WithContext(ctx)
			return true
		}
	}
	return false
}

type Option interface {
	Apply(opt *option)
}

type option struct {
	maxIdleConn int
	maxOpenConn int
	maxLifetime time.Duration
	DebugLog    bool
}

type optionFunc func(*option)

func (of optionFunc) Apply(cfg *option) { of(cfg) }

func ApplyOptions(opt *option, fns ...Option) {
	for _, f := range fns {
		f.Apply(opt)
	}
}

func WithMaxIdleConn(maxIdleConn int) Option {
	return optionFunc(func(opt *option) {
		opt.maxIdleConn = maxIdleConn
	})
}

func WithMaxOpenConn(maxOpenConn int) Option {
	return optionFunc(func(opt *option) {
		opt.maxOpenConn = maxOpenConn
	})
}

func WithMaxLifetime(maxLifetime time.Duration) Option {
	return optionFunc(func(opt *option) {
		opt.maxLifetime = maxLifetime
	})
}

func WithDebugLog(isDebug bool) Option {
	return optionFunc(func(opt *option) {
		opt.DebugLog = isDebug
	})
}

func Init(dbName string, dbConfig *Config, isDebug bool) error {
	err := dbConfig.Validate()
	if err != nil {
		return err
	}
	db, err := initDriver(dbConfig,
		WithMaxIdleConn(dbConfig.MaxIdleConnes),
		WithMaxOpenConn(dbConfig.MaxOpenConnes),
		WithDebugLog(isDebug),
		WithMaxLifetime(time.Duration(dbConfig.MaxLifetime)*time.Second),
	)
	if err != nil {
		return err
	}
	dbInstance, err := db.DB()
	if err != nil {
		return err
	}

	err = dbInstance.Ping()
	if err != nil {
		return err
	}

	configMap.Store(dbName, db)

	return nil
}

func initDriver(conf *Config, opts ...Option) (*gorm.DB, error) {
	opt := new(option)
	ApplyOptions(opt, opts...)

	gromConf := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: &Logger{},
	}
	if opt.DebugLog {
		gromConf.Logger.LogMode(gormLogger.Info)
	}
	driver, ok := GetDriver(conf.Driver)
	if !ok {
		return nil, errors.New("not supported database driver name:" + conf.Driver)
	}

	db, err := gorm.Open(driver.Dial(conf), gromConf)
	if err != nil {
		logger.Fatal(context.Background(), "database", err)
		return nil, err
	}
	err = db.Use(
		tracing.NewPlugin(
			tracing.WithoutMetrics(),
			tracing.WithAttributes(
				attribute.String("database_driver", conf.Driver),
				attribute.String("database_host", conf.Host),
				attribute.Int("database_port", conf.Port),
				attribute.String("database_db", conf.Database),
			),
		),
	)
	if err != nil {
		logger.Fatal(context.Background(), "database", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal(context.Background(), "database", err)
		return nil, err
	}

	sqlDB.SetMaxIdleConns(opt.maxIdleConn)
	sqlDB.SetMaxOpenConns(opt.maxOpenConn)
	sqlDB.SetConnMaxLifetime(opt.maxLifetime)
	return db, nil
}
