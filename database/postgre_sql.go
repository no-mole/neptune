package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	RegistryDriver("postgreSql", PostgreSqlDriver{})
}

type PostgreSqlDriver struct{}

func (p PostgreSqlDriver) Dial(conf *Config) gorm.Dialector {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		conf.Host,
		conf.Username,
		conf.Password,
		conf.Database,
		conf.Port,
	)
	return postgres.Open(dsn)
}
