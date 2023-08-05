package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/no-mole/neptune/config"
	"strconv"
	"strings"
)

type NaCosConfig struct {
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
	Endpoint  string `json:"endpoint" yaml:"endpoint"`
	Namespace string `json:"namespace" yaml:"namespace"`
}

type NaCosRegister struct {
	client naming_client.INamingClient
	errCh  chan error
}

func NewNaCosRegister(_ context.Context, conf *NaCosConfig, errCh chan error) (_ Registration, err error) {
	address := strings.Split(conf.Endpoint, ":")
	if len(address) < 2 {
		return nil, errors.New("illegal parameter of endpoint ")
	}
	port, err := strconv.ParseInt(address[1], 10, 64)
	if err != nil {
		return nil, err
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(address[0], uint64(port), constant.WithContextPath("/nacos")),
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(conf.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername(conf.Username),
		constant.WithPassword(conf.Password),
	)

	// create config client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, err
	}

	return &NaCosRegister{
		client: client,
		errCh:  errCh,
	}, nil
}

func (nacos *NaCosRegister) Register(_ context.Context, meta GrpcMeta) (err error) {
	success, err := nacos.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          config.GlobalConfig.IP,
		Port:        uint64(config.GlobalConfig.GrpcPort),
		ServiceName: meta.GenKey(),
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})
	if err != nil {
		nacos.errCh <- err
		return
	}
	if !success {
		nacos.errCh <- fmt.Errorf("注册实例%s失败", meta.GenKey())
		return
	}
	return
}

func (nacos *NaCosRegister) UnRegister(_ context.Context, meta GrpcMeta) (err error) {
	success, err := nacos.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          config.GlobalConfig.IP,
		Port:        uint64(config.GlobalConfig.GrpcPort),
		ServiceName: meta.GenKey(),
		Ephemeral:   true,
	})
	if err != nil {
		nacos.errCh <- err
		return
	}
	if !success {
		nacos.errCh <- fmt.Errorf("注册实例%s失败", meta.GenKey())
		return
	}
	return
}
