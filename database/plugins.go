package database

import "gorm.io/gorm"

type PluginFunc func(conf *Config) gorm.Plugin

var plugins = map[string]PluginFunc{}

func RegisterPlugin(name string, fn PluginFunc) {
	plugins[name] = fn
}

func PluginFuncByName(name string) (PluginFunc, bool) {
	fn, ok := plugins[name]
	return fn, ok
}
