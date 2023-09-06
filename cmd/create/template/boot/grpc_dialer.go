package boot

import (
	"context"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/grpc_dialer"
	"github.com/no-mole/neptune/grpc_service"
	middleware "github.com/no-mole/neptune/middlewares"
	"github.com/no-mole/neptune/protos/bar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Dialer(_ context.Context) application.HookFunc {
	return func(ctx context.Context) error {
		dialOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			middleware.OtelGrpcUnaryClientInterceptor(),
			middleware.OtelGrpcStreamClientInterceptor(),
		}
		return grpc_dialer.DialContext(
			ctx,
			dialOptions,
			grpc_service.ResolverNacosScheme,
			bar.Metadata,
			//... metadata
		)
	}
}
