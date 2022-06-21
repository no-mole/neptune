package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/no-mole/neptune/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

var manager *ServiceInstanceManager

func init() {
	manager = NewServiceInstanceManager()
}

type etcdResolver struct {
	scheme     string
	cli        *clientv3.Client
	clientConn resolver.ClientConn
	addrList   []resolver.Address
	once       sync.Once
}

func NewResolver(scheme string, client *clientv3.Client) (r resolver.Builder) {
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

//get addresses from local
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
	fmt.Println("resolve now")
	// TODO check
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
				r.clientConn.UpdateState(resolver.State{Addresses: r.addrList})
			}
			manager.AddNode(keyPrefix, &GrpcServiceInstance{Endpoint: addr})
		case clientv3.EventTypeDelete:
			if s, ok := remove(r.addrList, addr); ok {
				r.addrList = s
				r.clientConn.UpdateState(resolver.State{Addresses: r.addrList})
			}
			manager.DelNode(keyPrefix)
		}
	}
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
