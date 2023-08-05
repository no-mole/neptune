package registry

import (
	"context"
	"github.com/no-mole/neptune/config"
	"strings"
	"sync"
	"time"

	"github.com/no-mole/neptune/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type etcdResolver struct {
	scheme     string
	cli        *clientv3.Client
	clientConn resolver.ClientConn
	addrList   []resolver.Address
	once       sync.Once
}

func NewEtcdResolver(scheme string) (r resolver.Builder) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            strings.Split(config.GetRegistryConf().Endpoint, ","),
		Username:             config.GetRegistryConf().UserName,
		Password:             config.GetRegistryConf().Password,
		DialTimeout:          1 * time.Second,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: 1 * time.Second,
		PermitWithoutStream:  true,
	})
	if err != nil {
		return
	}
	return &etcdResolver{
		scheme: scheme,
		cli:    client,
	}
}

// Build etcd client
func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var err error
	keyPrefix := "/" + target.URL.Scheme + target.URL.Path + "/"
	r.clientConn = cc

	err = r.localCache(keyPrefix)
	if err != nil {
		return nil, err
	}

	r.once.Do(func() {
		go r.watch(keyPrefix)
	})
	return r, nil
}

// get addresses from local
func (r *etcdResolver) localCache(key string) error {
	instances := manager.Pick(key)
	if len(instances) == 0 {
		return nil
	}
	cacheList := make([]resolver.Address, len(instances))
	for i, v := range instances {
		cacheList[i] = resolver.Address{Addr: v.Endpoint}
	}
	return r.clientConn.UpdateState(resolver.State{Addresses: cacheList})
}

func (r *etcdResolver) Scheme() string {
	return r.scheme
}

// 当有连接被出现异常时，会触发该方法，因为这时候可能是有服务实例挂了，需要立即实现一次服务发现
func (r *etcdResolver) ResolveNow(rn resolver.ResolveNowOptions) {
	//无需操作，etcd,自动更新实例
}

// Close closes the resolver.
func (r *etcdResolver) Close() {
	logger.Info(context.Background(), "etcdResolver", logger.WithField("msg", "closed"))
}

func (r *etcdResolver) watch(keyPrefix string) {
	if r.cli == nil {
		return
	}

	getResp, err := r.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		logger.Error(context.Background(), "etcdResolver", err)
		return
	}

	for _, v := range getResp.Kvs {
		r.addrList = append(r.addrList, resolver.Address{
			Addr: string(v.Value),
		})
		manager.AddNode(keyPrefix, &GrpcServiceInstance{Endpoint: string(v.Value)})
	}
	r.clientConn.UpdateState(resolver.State{Addresses: r.addrList})

	watchChan := r.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			r.watchEvent(resp.Events, keyPrefix)
		}
	}
}

func (r *etcdResolver) watchEvent(events []*clientv3.Event, keyPrefix string) {
	for _, ev := range events {
		addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix)
		switch ev.Type {
		case clientv3.EventTypePut:
			if len(addr) == 0 {
				continue
			}
			if !exist(r.addrList, addr) {
				r.addrList = append(r.addrList, resolver.Address{Addr: addr})
				_ = r.clientConn.UpdateState(resolver.State{Addresses: r.addrList})
			}
			manager.AddNode(keyPrefix, &GrpcServiceInstance{Endpoint: addr})
		case clientv3.EventTypeDelete:
			if s, ok := remove(r.addrList, addr); ok {
				r.addrList = s
				_ = r.clientConn.UpdateState(resolver.State{Addresses: r.addrList})
			}
			manager.DelNode(keyPrefix, &GrpcServiceInstance{Endpoint: addr})
		}
	}
}
