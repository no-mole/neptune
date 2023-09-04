package middleware

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func OtelGrpcUnaryClientInterceptor() grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor())
}
func OtelGrpcStreamClientInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor())
}

func OtelGrpcUnaryServerInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor())
}
func OtelGrpcStreamServerInterceptor() grpc.ServerOption {
	return grpc.ChainStreamInterceptor(otelgrpc.StreamServerInterceptor())
}
