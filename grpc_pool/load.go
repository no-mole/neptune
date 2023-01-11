package grpc_pool

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/no-mole/neptune/config"
	clientv3 "go.etcd.io/etcd/client/v3"

	"google.golang.org/grpc/resolver"

	"github.com/no-mole/neptune/env"
	"github.com/no-mole/neptune/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

var pools map[string]Pool

func init() {
	pools = make(map[string]Pool)
}

const (
	// KeepAliveTime 在此时间后客户端没看到任何活动，将ping服务器
	KeepAliveTime = time.Duration(10) * time.Second

	// KeepAliveTimeout 客户端在ping以后等待的时间
	KeepAliveTimeout = time.Duration(3) * time.Second

	DefaultMaxIdle = 1

	DefaultMaxActive = 64

	DefaultMaxStreamsPerConn = 1000

	DefaultMaxConnIdleSecond = time.Minute

	DefaultMaxWaitConnTime = 20 * time.Millisecond
)

func WithOptions(opts ...Option) *Options {
	o := &Options{
		maxIdle:            DefaultMaxIdle,
		maxActive:          DefaultMaxActive,
		maxStreamsPerConn:  DefaultMaxStreamsPerConn,
		maxConnIdleSeconds: DefaultMaxConnIdleSecond,
		maxWaitConnTime:    DefaultMaxWaitConnTime,
		dialOptions:        []grpc.DialOption{},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func Init(ctx context.Context, opt *Options, connSetting ...registry.GrpcMeta) {
	cli, err := clientv3.New(clientv3.Config{
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

	for _, mm := range connSetting {
		meta := mm.Metadata()
		//todo scheme /
		scheme := fmt.Sprintf("%s%s", meta.Namespace, meta.Version)
		r := registry.NewResolver(scheme, cli)
		resolver.Register(r)
		pool, err := newPool(func() (*grpc.ClientConn, error) {
			return NodeDial(ctx, scheme,
				meta.ServiceName, opt)
		}, opt)
		if err != nil {
			return
		}
		pools[genPoolKey(mm)] = pool
	}
}

func NodeDial(ctx context.Context, scheme, serviceName string, opts *Options) (*grpc.ClientConn, error) {
	//retryOps := []grpc_retry.CallOption{
	//	grpc_retry.WithMax(2),
	//	grpc_retry.WithPerRetryTimeout(time.Second * 2),
	//	grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(time.Second, 0.2)),
	//}
	//retryInterceptor := grpc_retry.UnaryClientInterceptor(retryOps...)

	opts.dialOptions = append([]grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepAliveTime,
			Timeout:             KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}, opts.dialOptions...)
	//opts.dialOptions = append(opts.dialOptions, grpc.WithChainUnaryInterceptor(retryInterceptor))
	return grpc.DialContext(ctx, scheme+"://author/"+serviceName, opts.dialOptions...)
}

var ErrorKeyNotExist = errors.New("key not exist")

func GetConnection(meta registry.GrpcMeta) (Conn, error) {
	key := genPoolKey(meta)
	if pool, ok := pools[key]; ok {
		conn, err := pool.Get()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	return nil, ErrorKeyNotExist
}

func genPoolKey(m registry.GrpcMeta) string {
	meta := m.Metadata()
	return env.GetEnvMode() + "/" + meta.Version + "/" + meta.ServiceName
}

func Call(meta *registry.Metadata, fn func(conn grpc.ClientConnInterface) error) error {
	c, err := GetConnection(meta)
	if err != nil {
		return err
	}
	defer c.Close()
	return fn(c.Value())
}
