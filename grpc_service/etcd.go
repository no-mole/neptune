package grpc_service

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const ResolverEtcdScheme = "neptune-etcd"

// RegisterEtcdResolverBuilder 创建一个etcd解析器构建器，解析器schema为 ResolverEtcdScheme
func RegisterEtcdResolverBuilder(ctx context.Context, client *clientv3.Client, namespace string) resolver.Builder {
	builder := &EtcdResolverBuilder{
		ctx:       ctx,
		client:    client,
		namespace: namespace,
	}
	resolver.Register(builder)
	return builder
}

type EtcdResolverBuilder struct {
	ctx       context.Context
	client    *clientv3.Client
	namespace string
}

func (e *EtcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	// "neptune-etcd:///zeus/zeus.proto/zeus.ZeusService/v1"
	// target.Endpoints() = target.URL.Path = /zeus/zeus.proto/zeus.ZeusService/v1
	keyPrefix := target.Endpoint()
	if !strings.HasPrefix(keyPrefix, "/") {
		keyPrefix = "/" + keyPrefix
	}
	if !strings.HasSuffix(keyPrefix, "/") {
		keyPrefix = keyPrefix + "/"
	}
	keyPrefix = fmt.Sprintf("/%s/%s", e.namespace, keyPrefix)
	// keyPrefix = /namespace/zeus/zeus.proto/zeus.ZeusService/v1/
	// Item register endpoint = /namespace/zeus/zeus.proto/zeus.ZeusService/v1/192.168.1.1:8888
	newResolver := &etcdResolver{
		ctx:    e.ctx,
		client: e.client,
		prefix: keyPrefix,
		cc:     cc,
		close:  make(chan struct{}),
	}
	newResolver.start()
	return newResolver, nil
}

type etcdResolver struct {
	ctx    context.Context
	client *clientv3.Client
	prefix string
	cc     resolver.ClientConn
	opts   resolver.BuildOptions
	close  chan struct{}
}

func (e *EtcdResolverBuilder) Scheme() string {
	return ResolverEtcdScheme
}

func (e *etcdResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	getResp, err := e.client.Get(e.ctx, e.prefix, clientv3.WithPrefix())
	if err != nil {
		e.cc.ReportError(err)
		return
	}
	var address []resolver.Address
	for _, kv := range getResp.Kvs {
		endpoint := string(kv.Key[len(e.prefix):])
		_, _, err = net.SplitHostPort(endpoint)
		if err != nil {
			continue
		}
		address = append(address, resolver.Address{
			Addr: endpoint,
		})
	}
	if len(address) == 0 {
		return
	}
	err = e.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		e.cc.ReportError(err)
		return
	}
}

func (e *etcdResolver) Close() {
	close(e.close)
}

func (e *etcdResolver) start() {
	e.ResolveNow(resolver.ResolveNowOptions{})
	go func() {
		watchChan := e.client.Watch(e.ctx, e.prefix, clientv3.WithPrefix())
		for {
			select {
			case <-e.ctx.Done():
				return
			case _, ok := <-e.close:
				if !ok {
					return
				}
			case <-watchChan:
				e.ResolveNow(resolver.ResolveNowOptions{})
			}
		}
	}()
}

// NewEtcdRegister 创建一个etcd注册器
func NewEtcdRegister(ctx context.Context, client *clientv3.Client, namespace string, ttl int64) RegisterInterface {
	return &EtcdRegister{
		ctx:       ctx,
		client:    client,
		namespace: namespace,
		ttl:       ttl,
		services:  &RegisterServices{},
		close:     make(chan struct{}),
	}
}

type EtcdRegister struct {
	ctx context.Context

	namespace string
	client    *clientv3.Client

	ttl int64

	leaseID  clientv3.LeaseID
	services *RegisterServices

	close chan struct{}

	sync.Mutex
}

func (e *EtcdRegister) Close() error {
	close(e.close)
	return nil
}

func (e *EtcdRegister) Register(ctx context.Context, service Metadata, endpoint string) (err error) {
	e.Lock()
	defer e.Unlock()

	if e.leaseID == 0 {
		err = e.leaseKeepalive(ctx)
		if err != nil {
			return err
		}
	}

	_, err = e.client.Put(ctx, e.key(service, endpoint), e.value(), clientv3.WithLease(e.leaseID))

	if err != nil {
		return err
	}
	e.services.Put(service, endpoint)
	return err
}

func (e *EtcdRegister) Unregister(ctx context.Context, service Metadata, endpoint string) error {
	e.Lock()
	defer e.Unlock()
	_, err := e.client.Delete(ctx, e.key(service, endpoint))
	if err != nil {
		return err
	}
	e.services.Del(service, endpoint)
	return err
}

func (e *EtcdRegister) key(service Metadata, endpoint string) string {
	return fmt.Sprintf("/%s/%s/%s", e.namespace, service.UniqueKey(), endpoint)
}

func (e *EtcdRegister) value() string {
	name, _ := os.Hostname()
	return name
}

func (e *EtcdRegister) leaseKeepalive(ctx context.Context) error {
	grantResp, err := e.client.Grant(ctx, e.ttl)
	if err != nil {
		return err
	}

	e.leaseID = grantResp.ID

	ch, err := e.client.KeepAlive(ctx, e.leaseID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					e.reRegisterServices()
					return
				}
			case <-e.close:
				return
			case <-e.ctx.Done():
			}
		}
	}()

	return nil
}

func (e *EtcdRegister) reRegisterServices() {
	ticker := time.NewTicker(time.Duration(e.ttl) * time.Second)
	defer ticker.Stop()
	e.Lock()
	defer e.Unlock()

	for {
		err := e.leaseKeepalive(e.ctx)
		if err != nil {
			<-ticker.C
			continue
		}
		err = e.services.Range(func(instance Metadata, endpoint string) error {
			_, err := e.client.Put(e.ctx, e.key(instance, endpoint), e.value(), clientv3.WithLease(e.leaseID))
			return err
		})
		if err != nil {
			<-ticker.C
		} else {
			return
		}
	}
}

type RegisterServices struct {
	services map[string]*ServiceInfo
}

func (r *RegisterServices) Range(fn func(instance Metadata, endpoint string) error) error {
	for _, register := range r.services {
		for ep := range register.Endpoints {
			err := fn(register.Item, ep)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RegisterServices) Put(instance Metadata, endpoint string) {
	item, ok := r.services[instance.UniqueKey()]
	if !ok {
		item = &ServiceInfo{
			Item: instance,
		}
		r.services[instance.UniqueKey()] = item
	}
	item.Put(endpoint)
}

func (r *RegisterServices) Del(instance Metadata, endpoint string) {
	item, ok := r.services[instance.UniqueKey()]
	if !ok {
		return
	}
	item.Del(endpoint)
	if item.IsEmpty() {
		delete(r.services, instance.UniqueKey())
	}
}
