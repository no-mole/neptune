package registry

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/no-mole/neptune/config"
	"net/url"
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

	var serverConfigs []constant.ServerConfig
	for _, ep := range strings.Split(conf.Endpoint, ",") {
		scheme, host, port, path, err := getHostAndPort(ep)
		if err != nil {
			return nil, err
		}
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      host,
			ContextPath: path,
			Port:        port,
			Scheme:      scheme,
		})
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(conf.Namespace),
		constant.WithTimeoutMs(2000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("log/nacos_registry/log"),
		constant.WithCacheDir("log/nacos_registry/cache"),
		constant.WithLogLevel("warn"),
		constant.WithUsername(conf.Username),
		constant.WithPassword(conf.Password),
	)

	// create config client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: serverConfigs,
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

func getHostAndPort(rawurl string) (scheme, host string, port uint64, path string, err error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", "", 0, "", err
	}
	host = u.Hostname()
	portStr := u.Port()
	if portStr != "" {
		portInt, err := strconv.ParseUint(portStr, 10, 64)
		if err != nil {
			return "", "", 0, "", err
		}
		port = portInt
	} else if u.Scheme == "http" {
		port = 80
	} else if u.Scheme == "https" {
		port = 443
	}
	return u.Scheme, host, port, u.Path, nil
}
