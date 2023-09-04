package grpc_service

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/config"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"strconv"
)

// NewPlugin 服务注册、服务发现组件
func NewPlugin(_ context.Context) application.Plugin {
	plg := &Plugin{
		Plugin: application.NewPluginConfig("grpc-register", &application.PluginConfigOptions{
			ConfigName: "register",
			ConfigType: "yaml",
			EnvPrefix:  "register",
		}),
		config: &config.Config{},
	}
	plg.Command().PersistentFlags().StringVar(&plg.config.Type, "register-type", "", "config client type")
	plg.Command().PersistentFlags().StringVar(&plg.config.Endpoints, "register-endpoints", "", "config client endpoints")
	plg.Command().PersistentFlags().StringVar(&plg.config.Namespace, "register-namespace", "", "config client namespace")
	plg.Command().PersistentFlags().StringVar(&plg.config.Username, "register-username", "", "config client username")
	plg.Command().PersistentFlags().StringVar(&plg.config.Password, "register-password", "", "config client password")
	plg.Command().PersistentFlags().StringToStringVar(&plg.config.Settings, "register-settings", nil, "config client settings")

	return plg
}

type Plugin struct {
	application.Plugin
	config *config.Config
}

func (p *Plugin) Init(_ context.Context, config []byte) error {
	err := yaml.Unmarshal(config, p.config)
	if err != nil {
		return err
	}
	p.Command().RunE = func(cmd *cobra.Command, args []string) error {
		initFn, ok := registryClientTypes[p.config.Type]
		if !ok {
			return fmt.Errorf("unsupported register type:[%s]", p.config.Type)
		}
		return initFn(context.Background(), p.config)
	}
	err = p.Command().Execute()
	if err != nil {
		return err
	}
	p.Command().RunE = func(cmd *cobra.Command, args []string) error { return nil }
	return nil
}

type InitRegisterFunc func(ctx context.Context, conf *config.Config) error

var registryClientTypes = map[string]InitRegisterFunc{}

func RegistryClientType(typeName string, fn InitRegisterFunc) {
	registryClientTypes[typeName] = fn
}

var DefaultEtcdTTL int64 = 10

func init() {
	RegistryClientType("etcd", func(ctx context.Context, conf *config.Config) error {
		clientConf := config.Trans2EtcdConfig(ctx, conf)
		cli, err := clientv3.New(clientConf)
		if err != nil {
			return err
		}
		ttl := DefaultEtcdTTL
		if ttlStr, ok := conf.Settings["ttl"]; ok {
			ttlInt, _ := strconv.Atoi(ttlStr)
			if ttlInt > 0 {
				ttl = int64(ttlInt)
			}
		}
		SetDefaultRegister(NewEtcdRegister(ctx, cli, ttl))
		RegisterEtcdResolverBuilder(ctx, cli, ttl)
		return nil
	})
	RegistryClientType("nacos", func(ctx context.Context, conf *config.Config) error {
		groupName := ""
		if groupStr, ok := conf.Settings["groupName"]; ok {
			groupName = groupStr
		}
		clientConfig, serverConfigs, err := config.Trans2NacosConfig(ctx, conf)
		if err != nil {
			return err
		}

		// create nacos client
		cli, err := clients.NewNamingClient(
			vo.NacosClientParam{
				ClientConfig:  clientConfig,
				ServerConfigs: serverConfigs,
			},
		)
		if err != nil {
			return err
		}
		SetDefaultRegister(NewNacosRegister(ctx, cli, groupName))
		RegisterNacosResolverBuilder(ctx, cli)
		return nil
	})
}
