package center

import (
	"context"
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"strconv"
	"strings"
	"time"
)

type NaCos struct {
	opt     *config
	client  config_client.IConfigClient
	closeCh chan struct{}
}

func (s *NaCos) Init(opts ...Option) error {

	s.opt = ApplyOptions(opts...)

	if s.opt.Endpoint == "" {
		s.opt.Endpoint = defaultEtcdEndpoint
	}
	address := strings.Split(s.opt.Endpoint, ":")
	if len(address) < 2 {
		return errors.New("illegal parameter of endpoint ")
	}
	port, err := strconv.ParseInt(address[1], 10, 64)
	if err != nil {
		return err
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(address[0], uint64(port), constant.WithContextPath("/nacos")),
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(s.opt.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername(s.opt.Auth.Username),
		constant.WithPassword(s.opt.Auth.Password),
	)

	// create config client
	cli, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return err
	}
	s.client = cli
	return nil

}

func (s *NaCos) Close() error {
	close(s.closeCh)
	s.client.CloseClient()
	return nil
}

func (s *NaCos) Set(_ context.Context, key, value string) error {
	// 发布配置
	success, err := s.client.PublishConfig(
		vo.ConfigParam{
			DataId:  key,
			Group:   "DEFAULT_GROUP",
			Content: value},
	)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("set key %s fail", key)
	}
	return nil
}

func (s *NaCos) SetEX(_ context.Context, key, value string, ttl int64) error {
	success, err := s.client.PublishConfig(
		vo.ConfigParam{
			DataId:  "dataId",
			Group:   "DEFAULT_GROUP",
			Content: value,
			Type:    "json",
		})
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("set key %s fail", key)
	}
	go func() {
		<-time.After(time.Second * time.Duration(ttl))
		_, _ = s.client.DeleteConfig(vo.ConfigParam{
			DataId: "dataId",
			Group:  "",
		})
	}()
	return nil
}
func (s *NaCos) SetExKeepAlive(ctx context.Context, key, value string, ttl int64) error {
	return s.Set(ctx, key, value)
}

func (s *NaCos) Get(ctx context.Context, key string) (*Item, error) {
	value, err := s.client.GetConfig(
		vo.ConfigParam{
			DataId: key,
			Group:  "DEFAULT_GROUP",
		},
	)
	if err != nil {
		return nil, err
	}
	return &Item{
		Namespace: s.opt.Namespace,
		Key:       key,
		value:     value,
		IsDefault: false,
	}, nil
}
func (s *NaCos) GetDefault(ctx context.Context, key string, defaultValue string) (*Item, error) {
	value, err := s.client.GetConfig(
		vo.ConfigParam{
			DataId: key,
			Group:  "",
		},
	)
	if err != nil {
		return nil, err
	}
	item := &Item{
		Namespace: s.opt.Namespace,
		Key:       key,
		value:     value,
		IsDefault: false,
	}

	if value == "" {
		item.value = defaultValue
		item.IsDefault = true
	}

	return item, nil
}
func (s *NaCos) GetWithPrefixKey(ctx context.Context, prefixKey string) (*Item, error) {
	//nacos无法做前缀查询，只能做模糊搜索
	page, err := s.client.SearchConfig(vo.SearchConfigParam{
		Search: "blur",
		DataId: prefixKey,
	})
	if err != nil {
		return nil, err
	}
	if page.TotalCount == 0 {
		return nil, errors.New("record not found")
	}

	return &Item{
		Namespace: s.opt.Namespace,
		Key:       page.PageItems[0].DataId,
		value:     page.PageItems[0].Content,
		IsDefault: false,
	}, nil
}

func (s *NaCos) Exist(ctx context.Context, key string) (bool, error) {
	return true, nil //todo
}

func (s *NaCos) Watch(ctx context.Context, item *Item, callback func(item *Item)) {
	_ = s.client.ListenConfig(vo.ConfigParam{
		DataId: item.Key,
		Group:  "DEFAULT_GROUP",
		OnChange: func(namespace, group, dataId, data string) {
			callback(item)
		},
	})

}
func (s *NaCos) WatchWithPrefix(ctx context.Context, item *Item, callback func(item *Item)) {
	ret, err := s.GetWithPrefixKey(ctx, item.Key)
	if err != nil {
		return
	}
	s.Watch(ctx, ret, callback)
}
