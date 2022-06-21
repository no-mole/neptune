package model

import (
	"github.com/no-mole/neptune/env"
	"github.com/no-mole/neptune/json"
)

type App struct {
	BasePath   string
	ConfigPath string

	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	HttpPort  int    `yaml:"httpPort"`
	GrpcPort  int    `yaml:"grpcPort"`
	Namespace string `yaml:"namespace"` //命名空间

	Hostname string
	IP       string

	Env env.Env

	Registry *Registry

	Config *Config
}

func (a *App) String() string {
	data, _ := json.Marshal(a)
	return string(data)
}
