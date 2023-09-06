package database

import (
	"gorm.io/gorm"
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
