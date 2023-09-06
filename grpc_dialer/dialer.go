package grpc_dialer

import (
	"context"
	"fmt"
	"github.com/no-mole/neptune/grpc_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
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

func DialContext(ctx context.Context, opts []grpc.DialOption, scheme string, mds ...grpc_service.Metadata) error {
	newOpts := make([]grpc.DialOption, 0, len(DefaultDialOptions)+len(opts))
	newOpts = append(newOpts, DefaultDialOptions...)
	newOpts = append(newOpts, opts...)
	for _, md := range mds {
		cc, err := grpc.DialContext(ctx, fmt.Sprintf("%s://%s", scheme, md.UniqueKey()), newOpts...)
		if err != nil {
			return err
		}
		clientConnMapper[md.UniqueKey()] = cc
	}
	return nil
}

func Call(service grpc_service.Metadata, cb func(*grpc.ClientConn) error) error {
	cc, ok := clientConnMapper[service.UniqueKey()]
	if !ok {
		return fmt.Errorf("no client conn for service [%s]", service.UniqueKey())
	}
	return cb(cc)
}
