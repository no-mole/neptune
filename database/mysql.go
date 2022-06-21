package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	RegistryDriver("mysql", MysqlDriver{})
}

type MysqlDriver struct{}

func (m MysqlDriver) Dial(conf *Config) gorm.Dialector {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&timeout=2s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Database,
	)
	return mysql.Open(dsn)
}
