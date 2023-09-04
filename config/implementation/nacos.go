package implementation

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/no-mole/neptune/config"
)

var RegistryImplementationTypeNameNacos = "nacos"

func init() {
	config.RegistryImplementation(RegistryImplementationTypeNameNacos, func(ctx context.Context) config.Client {
		return &NacosConfigClient{}
	})
}

type NacosConfigClient struct {
	group   string
	conf    *config.Config
	client  config_client.IConfigClient
	closeCh chan struct{}
}

func (s *NacosConfigClient) Init(ctx context.Context, conf *config.Config) error {
	s.conf = conf
	s.group = "DEFAULT_GROUP"

	if group, ok := conf.Settings["group"]; ok {
		s.group = group
	}

	clientConfig, serverConfigs, err := config.Trans2NacosConfig(ctx, conf)
	if err != nil {
		return err
	}

	// create nacos client
	cli, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return err
	}

	s.client = cli
	return nil

}

func (s *NacosConfigClient) Close() error {
	close(s.closeCh)
	s.client.CloseClient()
	return nil
}

func (s *NacosConfigClient) Set(_ context.Context, key, value string) error {
	// 发布配置
	success, err := s.client.PublishConfig(
		vo.ConfigParam{
			DataId:  key,
			Group:   s.group,
			Content: value,
		},
	)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("set key %s fail", key)
	}
	return nil
}

func (s *NacosConfigClient) Get(ctx context.Context, key string) (*config.Item, error) {
	value, err := s.client.GetConfig(
		vo.ConfigParam{
			DataId: key,
			Group:  s.group,
		},
	)
	if err != nil {
		return nil, err
	}
	return config.NewItem(s.conf.Namespace, key, value), nil
}

func (s *NacosConfigClient) Exist(ctx context.Context, key string) (bool, error) {
	value, err := s.client.GetConfig(
		vo.ConfigParam{
			DataId: key,
			Group:  s.group,
		},
	)
	return value != "", err
}

func (s *NacosConfigClient) Watch(ctx context.Context, item *config.Item, callback func(item *config.Item)) error {
	return s.client.ListenConfig(vo.ConfigParam{
		DataId: item.Key,
		Group:  s.group,
		OnChange: func(namespace, group, dataId, data string) {
			item.SetValue(data)
			callback(item)
		},
	})
}
