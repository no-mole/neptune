package grpc_dialer

import (
	"context"
	"fmt"
	"time"

	"github.com/no-mole/neptune/grpc_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	// KeepAliveTime 在此时间后客户端没看到任何活动，将ping服务器
	KeepAliveTime = time.Duration(10) * time.Second

	// KeepAliveTimeout 客户端在ping以后等待的时间
	KeepAliveTimeout = time.Duration(3) * time.Second
)

var DefaultDialOptions = []grpc.DialOption{
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                KeepAliveTime,
		Timeout:             KeepAliveTimeout,
		PermitWithoutStream: true,
	}),
}
var clientConnMapper = map[string]*grpc.ClientConn{}

// DialContext 根据metadata构建链接池
func DialContext(ctx context.Context, opts []grpc.DialOption, scheme string, mds ...grpc_service.Metadata) error {
	opts = mergeDialOpts(DefaultDialOptions, opts)
	for _, md := range mds {
		cc, err := grpc.DialContext(ctx, fmt.Sprintf("%s://%s", scheme, md.UniqueKey()), opts...)
		if err != nil {
			return err
		}
		clientConnMapper[md.UniqueKey()] = cc
	}
	return nil
}

// DialContextConnection 根据metadata构建单个链接
func DialContextConnection(ctx context.Context, opts []grpc.DialOption, scheme string, md grpc_service.Metadata) (*grpc.ClientConn, error) {
	opts = mergeDialOpts(DefaultDialOptions, opts)
	cc, err := grpc.DialContext(ctx, fmt.Sprintf("%s://%s", scheme, md.UniqueKey()), opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// DialContextEndpointConnection 使用固定端点构建链接
func DialContextEndpointConnection(ctx context.Context, opts []grpc.DialOption, endpoint string) (*grpc.ClientConn, error) {
	opts = mergeDialOpts(DefaultDialOptions, opts)
	cc, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// DialContextEndpoint 使用固定端点构建链接池
func DialContextEndpoint(ctx context.Context, opts []grpc.DialOption, endpoint string, mds ...grpc_service.Metadata) error {
	opts = mergeDialOpts(DefaultDialOptions, opts)
	cc, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return err
	}
	for _, md := range mds {
		clientConnMapper[md.UniqueKey()] = cc
	}
	return nil
}

func DialContextConn(conn *grpc.ClientConn, mds ...grpc_service.Metadata) error {
	for _, md := range mds {
		clientConnMapper[md.UniqueKey()] = conn
	}
	return nil
}

func Call(md grpc_service.Metadata, cb func(*grpc.ClientConn) error) error {
	cc, ok := clientConnMapper[md.UniqueKey()]
	if !ok {
		return fmt.Errorf("no client conn for md [%s]", md.UniqueKey())
	}
	return cb(cc)
}

func mergeDialOpts(opts1, opts2 []grpc.DialOption) []grpc.DialOption {
	newOpts := make([]grpc.DialOption, 0, len(opts1)+len(opts2))
	newOpts = append(newOpts, DefaultDialOptions...)
	newOpts = append(newOpts, opts2...)
	return newOpts
}
