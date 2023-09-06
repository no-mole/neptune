package grpc_service

import (
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc/resolver"
	"net"
	"strconv"
	"strings"
)

const ResolverNacosScheme = "neptune-nacos"

// RegisterNacosResolverBuilder 创建一个nacos解析器构建器，解析器schema为 ResolverEtcdScheme
func RegisterNacosResolverBuilder(ctx context.Context, client naming_client.INamingClient) resolver.Builder {
	builder := &NacosResolverBuilder{
		ctx:    ctx,
		client: client,
		close:  make(chan struct{}),
	}
	resolver.Register(builder)
	return builder
}

type NacosResolverBuilder struct {
	ctx context.Context

	client naming_client.INamingClient

	close chan struct{}
}

func (n *NacosResolverBuilder) RegisterService(_ context.Context, service Metadata, endpoint string) error {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	success, err := n.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(portInt),
		ServiceName: service.UniqueKey(),
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("register service [%s] unsuccess in %s", service.UniqueKey(), n.Scheme())
	}
	return nil
}

func (n *NacosResolverBuilder) UnregisterService(_ context.Context, service Metadata, endpoint string) error {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	success, err := n.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(portInt),
		ServiceName: service.UniqueKey(),
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("unregister service [%s] unsuccess in %s", service.UniqueKey(), n.Scheme())
	}
	return nil
}

func (n *NacosResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// "neptune-nacos:///namespace/zeus/zeus.proto/zeus.ZeusService/v1"
	// target.Endpoints() = target.URL.Path = namespace/zeus/zeus.proto/zeus.ZeusService/v1
	key := target.Endpoint()
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	// key = /namespace/zeus/zeus.proto/zeus.ZeusService/v1/
	// Item register endpoint = /namespace/zeus/zeus.proto/zeus.ZeusService/v1/192.168.1.1:8888
	instance := &nacosResolver{
		ctx:    n.ctx,
		client: n.client,
		close:  n.close,
		key:    key,
		cc:     cc,
	}
	instance.start()
	return instance, nil
}

func (n *NacosResolverBuilder) Scheme() string {
	return ResolverNacosScheme
}

type nacosResolver struct {
	ctx    context.Context
	client naming_client.INamingClient
	close  chan struct{}
	key    string
	cc     resolver.ClientConn
	opts   resolver.BuildOptions
}

func (n *nacosResolver) ResolveNow(options resolver.ResolveNowOptions) {
	// 获取服务列表
	param := vo.SelectInstancesParam{
		ServiceName: n.key,
		HealthyOnly: true,
	}
	instances, err := n.client.SelectInstances(param)
	if err != nil {
		n.cc.ReportError(err)
		return
	}
	var address []resolver.Address
	for _, v := range instances {
		if !v.Enable {
			continue
		}
		address = append(address, resolver.Address{
			Addr: fmt.Sprintf("%s:%d", v.Ip, v.Port),
		})
	}
	err = n.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		n.cc.ReportError(err)
		return
	}
}

func (n *nacosResolver) Close() {
	close(n.close)
}

func (n *nacosResolver) start() {
	n.ResolveNow(resolver.ResolveNowOptions{})
	go func() {
		_ = n.client.Subscribe(&vo.SubscribeParam{
			ServiceName: n.key,
			SubscribeCallback: func(services []model.Instance, err error) {
				if err != nil {
					n.cc.ReportError(err)
				}
				var address []resolver.Address
				// 处理实例列表
				for _, v := range services {
					if !v.Enable {
						continue
					}
					address = append(address, resolver.Address{
						Addr: fmt.Sprintf("%s:%d", v.Ip, v.Port),
					})
				}
				err = n.cc.UpdateState(resolver.State{
					Addresses: address,
				})
				if err != nil {
					n.cc.ReportError(err)
				}
			},
		})
	}()
}

// NewNacosRegister 创建一个nacos服务注册器
func NewNacosRegister(ctx context.Context, client naming_client.INamingClient, groupName string) RegisterInterface {
	return &NacosRegister{
		ctx:       ctx,
		client:    client,
		groupName: groupName,
	}
}

type NacosRegister struct {
	ctx       context.Context
	client    naming_client.INamingClient
	groupName string
}

func (n *NacosRegister) Close() error {
	return nil
}

func (n *NacosRegister) Register(_ context.Context, service Metadata, endpoint string) error {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	success, err := n.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(portInt),
		ServiceName: service.UniqueKey(),
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		GroupName:   n.groupName,
	})
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("nacos register: register service [%s] unsuccess", service.UniqueKey())
	}
	return nil
}

func (n *NacosRegister) Unregister(_ context.Context, service Metadata, endpoint string) error {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	success, err := n.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(portInt),
		ServiceName: service.UniqueKey(),
		Ephemeral:   true,
		GroupName:   n.groupName,
	})
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("naocs register: unregister service [%s] unsuccess", service.UniqueKey())
	}
	return nil
}
