package grpc_dialer

import (
	"context"

	"google.golang.org/grpc"

	"github.com/no-mole/neptune/grpc_service"
)

var Conn = map[string]*grpc.ClientConn{}

type GrpcDialer struct {
	metadataInfo map[string][]grpc_service.Metadata
	opts         []grpc.DialOption
	endpoints    map[string]string
}

func NewGrpcDialer(
	endpoints map[string]string,
	metadataInfo map[string][]grpc_service.Metadata,
	opt []grpc.DialOption) *GrpcDialer {
	g := new(GrpcDialer)
	g.endpoints = endpoints
	g.opts = opt
	g.metadataInfo = metadataInfo
	return g
}

func (g *GrpcDialer) Run(ctx context.Context) error {
	opts := mergeDialOpts(DefaultDialOptions, g.opts)
	for service, metadataInfos := range g.metadataInfo {
		if endpoint, exist := g.endpoints[service]; exist {
			cc, err := grpc.DialContext(ctx, endpoint, opts...)
			if err != nil {
				return err
			}

			err = DialContextConn(cc, metadataInfos...)
			if err != nil {
				return err
			}
			Conn[service] = cc
		}
	}
	return nil
}
