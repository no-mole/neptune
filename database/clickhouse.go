package database

import (
	"fmt"
	"time"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func init() {
	RegistryDriver("clickhouse", ClickhouseDriver{})
}

type ClickhouseDriver struct{}

func (c ClickhouseDriver) Dial(conf *Config) gorm.Dialector {
	dsn := fmt.Sprintf("tcp://%s:%d?database=%s&username=%s&password=%s&read_timeout=%d&write_timeout=%d",
		conf.Host,
		conf.Port,
		conf.Database,
		conf.Username,
		conf.Password,
		conf.ReadTimeout*time.Millisecond,
		conf.WriteTimeout*time.Millisecond,
	)
	return clickhouse.Open(dsn)
}
