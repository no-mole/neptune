package config

import (
	"context"
	"errors"
	"github.com/no-mole/neptune/application"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewConfigCenterPlugin 配置中心组件
func NewConfigCenterPlugin() application.Plugin {
	configOpts := &application.PluginConfigOptions{
		ConfigName: "config",
		ConfigType: "yaml",
		EnvPrefix:  "config",
	}
	plg := &Plugin{
		Plugin: application.NewPluginConfig("config", configOpts),
		config: &Config{},
	}

	plg.Command().PersistentFlags().StringVar(&plg.config.Type, "config-type", "", "config client type")
	plg.Command().PersistentFlags().StringVar(&plg.config.Endpoints, "config-endpoints", "", "config client endpoints")
	plg.Command().PersistentFlags().StringVar(&plg.config.Namespace, "config-namespace", "", "config client namespace")
	plg.Command().PersistentFlags().StringVar(&plg.config.Username, "config-username", "", "config client username")
	plg.Command().PersistentFlags().StringVar(&plg.config.Password, "config-password", "", "config client password")
	plg.Command().PersistentFlags().StringToStringVar(&plg.config.Settings, "config-settings", nil, "config client settings")

	return plg
}

type Plugin struct {
	application.Plugin
	config *Config
}

func (p *Plugin) Init(_ context.Context, conf []byte) error {
	err := yaml.Unmarshal(conf, p.config)
	if err != nil {
		return err
	}
	p.Command().RunE = func(cmd *cobra.Command, args []string) error {
		if p.config.Type == "" {
			return errors.New("config plugin used but not initialization")
		}
		return Init(context.Background(), p.config)
	}
	err = p.Command().Execute()
	if err != nil {
		return err
	}
	p.Command().RunE = func(cmd *cobra.Command, args []string) error { return nil }
	return nil
}
