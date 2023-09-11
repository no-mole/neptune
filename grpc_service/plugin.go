package grpc_service

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/config"
	"github.com/no-mole/neptune/logger"
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
			EnvPrefix:  "",
		}),
		config: &config.Config{},
	}
	plg.Flags().StringVar(&plg.config.Type, "register-type", "", "config client type")
	plg.Flags().StringVar(&plg.config.Endpoints, "register-endpoints", "", "config client endpoints")
	plg.Flags().StringVar(&plg.config.Namespace, "register-namespace", "", "config client namespace")
	plg.Flags().StringVar(&plg.config.Username, "register-username", "", "config client username")
	plg.Flags().StringVar(&plg.config.Password, "register-password", "", "config client password")
	plg.Flags().StringToStringVar(&plg.config.Settings, "register-settings", nil, "config client settings")
	return plg
}

type Plugin struct {
	application.Plugin
	config *config.Config
}

func (p *Plugin) Config(ctx context.Context, conf []byte) error {
	return yaml.Unmarshal(conf, p.config)
}

func (p *Plugin) Init(ctx context.Context) error {
	logger.Info(
		ctx,
		"config center plugin init",
		logger.WithField("grpcRegisterType", p.config.Type),
		logger.WithField("grpcRegisterEndpoints", p.config.Endpoints),
		logger.WithField("grpcRegisterNamespace", p.config.Namespace),
		logger.WithField("grpcRegisterUsername", p.config.Username),
		logger.WithField("grpcRegisterSettings", p.config.Settings),
	)
	initFn, ok := registryClientTypes[p.config.Type]
	if !ok {
		return fmt.Errorf("unsupported register type:[%s]", p.config.Type)
	}
	return initFn(context.Background(), p.config)
}
func (p *Plugin) Run(ctx context.Context) error {
	<-ctx.Done()
	return Close()
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
		SetDefaultRegister(NewEtcdRegister(ctx, cli, conf.Namespace, ttl))
		RegisterEtcdResolverBuilder(ctx, cli, conf.Namespace)
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
