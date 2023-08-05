package registry

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/no-mole/neptune/config"
	"github.com/no-mole/neptune/logger"
	"google.golang.org/grpc/resolver"
	"strconv"
	"strings"
	"sync"
)

type naCosResolver struct {
	scheme     string
	clientConn resolver.ClientConn
	addrList   []resolver.Address
	once       sync.Once
	client     naming_client.INamingClient
}

func NewNaCosResolver(scheme string) (r resolver.Builder) {

	endpoints := strings.Split(config.GetRegistryConf().Endpoint, ",")
	sc := make([]constant.ServerConfig, 0, len(endpoints))
	for _, v := range endpoints {
		address := strings.Split(v, ":")

		if len(address) < 2 {
			return
		}
		port, err := strconv.ParseInt(address[1], 10, 64)
		if err != nil {
			return
		}
		sc = append(sc, *constant.NewServerConfig(address[0], uint64(port), constant.WithContextPath("/nacos")))

	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(config.GlobalConfig.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername(config.GetRegistryConf().UserName),
		constant.WithPassword(config.GetRegistryConf().Password),
	)

	// create config client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return
	}
	return &naCosResolver{
		scheme: scheme,
		client: client,
	}
}

func (nacos *naCosResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var err error

	///neptune/v1/role.RoleService
	dataId := "/" + target.URL.Scheme + target.URL.Path + "/"
	nacos.clientConn = cc

	err = nacos.localCache(dataId)
	if err != nil {
		return nil, err
	}

	nacos.once.Do(func() {
		go nacos.watch(dataId)
	})
	return nacos, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (nacos *naCosResolver) Scheme() string {
	return nacos.scheme
}

func (nacos *naCosResolver) watch(serviceName string) {
	// 获取服务列表
	param := vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
	}
	instances, err := nacos.client.SelectInstances(param)
	if err != nil {
		logger.Error(context.Background(), "naCosResolver", err)
		return
	}
	for _, v := range instances {
		nacos.addrList = append(nacos.addrList, resolver.Address{
			Addr: fmt.Sprintf("%s:%d", v.Ip, v.Port),
		})
		_ = nacos.clientConn.UpdateState(resolver.State{Addresses: nacos.addrList})
		manager.AddNode(serviceName, &GrpcServiceInstance{Endpoint: fmt.Sprintf("%s:%d", v.Ip, v.Port)})
	}

	err = nacos.client.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		SubscribeCallback: func(services []model.Instance, err error) {
			// 处理实例列表
			for _, v := range services {
				addr := fmt.Sprintf("%s:%d", v.Ip, v.Port)
				if !v.Enable && exist(nacos.addrList, addr) {
					if s, ok := remove(nacos.addrList, addr); ok {
						nacos.addrList = s
						_ = nacos.clientConn.UpdateState(resolver.State{Addresses: nacos.addrList})
					}
					manager.DelNode(serviceName, &GrpcServiceInstance{Endpoint: addr})
				}

				if !exist(nacos.addrList, addr) {
					nacos.addrList = append(nacos.addrList, resolver.Address{Addr: addr})
					_ = nacos.clientConn.UpdateState(resolver.State{Addresses: nacos.addrList})
				}
				manager.AddNode(serviceName, &GrpcServiceInstance{Endpoint: addr})

			}
		},
	})
	if err != nil {
		logger.Error(context.Background(), "naCosResolver", err)
		return
	}

}

func (nacos *naCosResolver) ResolveNow(rn resolver.ResolveNowOptions) {
	//无需操作，nacos会watch,自动更新实例
}

func (nacos *naCosResolver) localCache(dataId string) error {
	instances := manager.Pick(dataId)
	if len(instances) == 0 {
		return nil
	}
	cacheList := make([]resolver.Address, len(instances))
	for i, v := range instances {
		cacheList[i] = resolver.Address{Addr: v.Endpoint}
	}
	return nacos.clientConn.UpdateState(resolver.State{Addresses: cacheList})
}

// Close closes the resolver.
func (nacos *naCosResolver) Close() {
	logger.Info(context.Background(), "nacosRegister", logger.WithField("msg", "closed"))
}
