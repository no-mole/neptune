package config

import (
	"context"
	"errors"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/logger"
	"gopkg.in/yaml.v3"
)

// NewConfigCenterPlugin 配置中心组件
func NewConfigCenterPlugin(_ context.Context) application.Plugin {
	configOpts := &application.PluginConfigOptions{
		ConfigName: "config.yaml",
		ConfigType: "yaml",
		EnvPrefix:  "",
	}
	plg := &Plugin{
		Plugin: application.NewPluginConfig("config-center", configOpts),
		config: &Config{},
	}

	plg.Flags().StringVar(&plg.config.Type, "config-type", "", "config client type")
	plg.Flags().StringVar(&plg.config.Endpoints, "config-endpoints", "", "config client endpoints")
	plg.Flags().StringVar(&plg.config.Namespace, "config-namespace", "", "config client namespace")
	plg.Flags().StringVar(&plg.config.Username, "config-username", "", "config client username")
	plg.Flags().StringVar(&plg.config.Password, "config-password", "", "config client password")
	plg.Flags().StringToStringVar(&plg.config.Settings, "config-settings", nil, "config client settings")
	return plg
}

type Plugin struct {
	application.Plugin
	config *Config
}

func (p *Plugin) Config(_ context.Context, conf []byte) error {
	return yaml.Unmarshal(conf, p.config)
}

func (p *Plugin) Init(ctx context.Context) error {
	logger.Info(
		ctx,
		"config center plugin init",
		logger.WithField("configCenterType", p.config.Type),
		logger.WithField("configCenterEndpoints", p.config.Endpoints),
		logger.WithField("configCenterNamespace", p.config.Namespace),
		logger.WithField("configCenterUsername", p.config.Username),
		logger.WithField("configCenterSettings", p.config.Settings),
	)
	if p.config.Type == "" {
		return errors.New("config plugin used but not initialization")
	}
	return InitDefaultClient(context.Background(), p.config)
}

func (p *Plugin) Run(ctx context.Context) error {
	<-ctx.Done()
	return defaultClient.Close()
}
