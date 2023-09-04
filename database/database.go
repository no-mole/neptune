package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	validate "github.com/go-playground/validator/v10"
	"github.com/no-mole/neptune/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	//databases map[string]*gorm.DB
	databases sync.Map
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

	Plugins []string `json:"plugins" yaml:"plugins"`
}

func (c *Config) Validate() error {
	return validator.Struct(c)
}

type BaseDb struct {
	*gorm.DB
}

func (b *BaseDb) SetEngine(ctx context.Context, engine string) bool {
	if value, ok := databases.Load(engine); ok {
		if db, ok := value.(*gorm.DB); ok {
			b.DB = db.WithContext(ctx)
			return true
		}
	}
	return false
}

type option struct {
	maxIdleConn  int
	maxOpenConn  int
	maxLifetime  time.Duration
	DebugLog     bool
	PluginsNames []string
}

type OptionFunc func(*option)

func ApplyOptions(opt *option, fns ...OptionFunc) {
	for _, f := range fns {
		f(opt)
	}
}

func WithPlugins(plugins ...string) OptionFunc {
	return func(o *option) {
		o.PluginsNames = append(o.PluginsNames, plugins...)
	}
}

func WithMaxIdleConn(maxIdleConn int) OptionFunc {
	return func(opt *option) {
		opt.maxIdleConn = maxIdleConn
	}
}

func WithMaxOpenConn(maxOpenConn int) OptionFunc {
	return func(opt *option) {
		opt.maxOpenConn = maxOpenConn
	}
}

func WithMaxLifetime(maxLifetime time.Duration) OptionFunc {
	return func(opt *option) {
		opt.maxLifetime = maxLifetime
	}
}

func WithDebugLog(isDebug bool) OptionFunc {
	return func(opt *option) {
		opt.DebugLog = isDebug
	}
}

func Init(dbName string, conf *Config, isDebug bool) error {
	err := conf.Validate()
	if err != nil {
		return err
	}

	db, err := initDriver(conf,
		WithMaxIdleConn(conf.MaxIdleConnes),
		WithMaxOpenConn(conf.MaxOpenConnes),
		WithDebugLog(isDebug),
		WithMaxLifetime(time.Duration(conf.MaxLifetime)*time.Second),
		WithPlugins(),
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

	databases.Store(dbName, db)

	return nil
}

func initDriver(conf *Config, opts ...OptionFunc) (*gorm.DB, error) {
	opt := new(option)
	ApplyOptions(opt, opts...)

	gormConf := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: NewLogger(
			logger.WithField("database_driver", conf.Driver),
			logger.WithField("database_host", conf.Host),
			logger.WithField("database_port", conf.Port),
			logger.WithField("database_database_name", conf.Database),
			logger.WithField("database_username", conf.Username),
		),
		Plugins: map[string]gorm.Plugin{},
	}

	for _, pluginName := range opt.PluginsNames {
		pluginFunc, ok := PluginFuncByName(pluginName)
		if !ok {
			return nil, fmt.Errorf("database plugin [%s] not registered", pluginName)
		}
		plugin := pluginFunc(conf)
		gormConf.Plugins[plugin.Name()] = plugin
	}

	if opt.DebugLog {
		gormConf.Logger.LogMode(gormLogger.Info)
	}

	driver, ok := GetDriver(conf.Driver)
	if !ok {
		return nil, errors.New("not supported database driver name:" + conf.Driver)
	}

	db, err := gorm.Open(driver.Dial(conf), gormConf)
	if err != nil {
		logger.Fatal(context.Background(), "database", err)
		return nil, err
	}

	instance, err := db.DB()
	if err != nil {
		logger.Fatal(context.Background(), "database", err)
		return nil, err
	}

	instance.SetMaxIdleConns(opt.maxIdleConn)
	instance.SetMaxOpenConns(opt.maxOpenConn)
	instance.SetConnMaxLifetime(opt.maxLifetime)
	return db, nil
}
