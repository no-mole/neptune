package config

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/no-mole/neptune/env"
	"gopkg.in/yaml.v2"

	"github.com/no-mole/neptune/utils"

	"github.com/no-mole/neptune/config/model"
)

var GlobalConfig *model.App

func Init(ctx context.Context) error {
	var err error
	rootDir := utils.GetCurrentAbPath()
	//init env
	GlobalConfig = &model.App{
		BasePath: rootDir,
		Env:      env.LoadEnv(rootDir),
		Registry: new(model.Registry),
		Config:   new(model.Config),
	}
	GlobalConfig.ConfigPath = path.Join(rootDir, "config", GlobalConfig.Env.Mode)

	loader := []func() error{loadIP, loadFile, loadEnv}
	for _, fn := range loader {
		err = fn()
		if err != nil {
			return err
		}
	}

	if GlobalConfig.Env.Debug {
		fmt.Printf("%s\n", GlobalConfig)
	}

	err = initCenterClient()
	if err != nil {
		return err
	}
	return nil
}

func loadIP() error {
	var err error
	//load ip
	GlobalConfig.IP, err = utils.GetSystemIP()
	if err != nil {
		return err
	}
	//load hostname
	GlobalConfig.Hostname, _ = os.Hostname()
	if GlobalConfig.Hostname == "" {
		GlobalConfig.Hostname = GlobalConfig.IP
	}
	return nil
}

func loadEnv() error {
	env.Load(
		&env.Item{
			Key: "GRPC_PORT",
			Setter: func(value string) {
				if GlobalConfig.GrpcPort == 0 {
					GlobalConfig.GrpcPort, _ = utils.GetAvailablePort()
				}
				port, _ := strconv.Atoi(value)
				if port > 0 {
					GlobalConfig.GrpcPort = port
				}
			},
		},
		&env.Item{
			Key: "HTTP_PORT",
			Setter: func(value string) {
				if value != "" {
					GlobalConfig.HttpPort, _ = strconv.Atoi(value)
				}
			},
		},
		&env.Item{
			Key: "REGISTRY_TYPE",
			Setter: func(value string) {
				if GlobalConfig.Registry.Type == "" {
					GlobalConfig.Registry.Type = "etcd"
				}
				if value != "" {
					GlobalConfig.Registry.Type = value
				}
			},
		},
		&env.Item{
			Key: "REGISTRY_ENDPOINT",
			Setter: func(value string) {
				if GlobalConfig.Registry.Endpoint == "" {
					GlobalConfig.Registry.Endpoint = "127.0.0.1:2379"
				}
				if value != "" {
					GlobalConfig.Registry.Endpoint = value
				}
			},
		},
		&env.Item{
			Key: "CONFIG_TYPE",
			Setter: func(value string) {
				if GlobalConfig.Config.Type == "" {
					GlobalConfig.Config.Type = "etcd"
				}
				if value != "" {
					GlobalConfig.Config.Type = value
				}
			},
		},
		&env.Item{
			Key: "CONFIG_ENDPOINT",
			Setter: func(value string) {
				if GlobalConfig.Config.Endpoint == "" {
					GlobalConfig.Config.Endpoint = "127.0.0.1:2379"
				}
				if value != "" {
					GlobalConfig.Config.Endpoint = value
				}
			},
		},
	)
	return nil
}

func loadFile() error {
	baseConfigDir := fmt.Sprintf("%s/config/%s", GlobalConfig.BasePath, GlobalConfig.Env.Mode)
	needLoad := []struct {
		path string
		out  interface{}
	}{
		{
			path: fmt.Sprintf("%s/config/%s", GlobalConfig.BasePath, "app.yaml"),
			out:  GlobalConfig,
		},
		{
			path: fmt.Sprintf("%s/registry.yaml", baseConfigDir),
			out:  &GlobalConfig.Registry,
		},
		{
			path: fmt.Sprintf("%s/config.yaml", baseConfigDir),
			out:  &GlobalConfig.Config,
		},
	}

	for _, item := range needLoad {
		exist := utils.FileExist(item.path)
		if exist {
			body, err := os.ReadFile(item.path)
			if err != nil {
				return err
			}
			err = yaml.Unmarshal(body, item.out)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetCenterConf 获取配置中心配置
func GetCenterConf() *model.Config {
	return GlobalConfig.Config
}

//GetRegistryConf 获取注册中心配置
func GetRegistryConf() *model.Registry {
	return GlobalConfig.Registry
}
