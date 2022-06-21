package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/tracing"
	"google.golang.org/grpc"
)

func startTracingFromMetadata(ctx context.Context, spanId string) context.Context {
	//从context取metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.Pairs()
	}
	//判断metadata是否有tracing的key，有则解析成tracing实体并且注入到context中
	if tracingMdArr := md[strings.ToLower(tracing.TracingContextKey)]; tracingMdArr != nil && len(tracingMdArr) > 0 {
		t, err := tracing.Decoding(tracingMdArr[0])
		if err == nil {
			ctx = tracing.WithContext(ctx, t)
		}
	}
	//开启新的tracing span
	ctx = tracing.Start(ctx, spanId)
	return ctx
}

func log(ctx context.Context, start, end time.Time, caller, errorMsg string) {
	if errorMsg == "" {
		logger.Info(
			ctx,
			"tracing",
			logger.WithField("start_time", start.Format(time.RFC3339)),
			logger.WithField("end_time", end.Format(time.RFC3339)),
			logger.WithField("duration", time.Since(start).Milliseconds()),
			logger.WithField("host", ""),
			logger.WithField("url", ""),
			logger.WithField("method", "grpc"),
			logger.WithField("caller", caller),
			logger.WithField("error", errorMsg),
		)
	} else {
		logger.Error(
			ctx,
			"tracing",
			errors.New(errorMsg),
			logger.WithField("start_time", start.Format(time.RFC3339)),
			logger.WithField("end_time", end.Format(time.RFC3339)),
			logger.WithField("duration", time.Since(start).Milliseconds()),
			logger.WithField("host", ""),
			logger.WithField("url", ""),
			logger.WithField("method", "grpc"),
			logger.WithField("caller", caller),
			logger.WithField("error", errorMsg),
		)
	}
}

// TracingServerInterceptor 拦截器
func TracingServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = startTracingFromMetadata(ctx, info.FullMethod)
		start := time.Now()
		errorMsg := ""
		defer func() {
			err := recover()
			if err != nil {
				errorMsg = fmt.Sprintf("%+v", err)
			}
			log(ctx, start, time.Now(), info.FullMethod, errorMsg)
		}()
		resp, err = handler(ctx, req)
		if err != nil {
			errorMsg = err.Error()
		}
		return resp, err
	}
}

func TracingServerStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := startTracingFromMetadata(ss.Context(), info.FullMethod)
		start := time.Now()
		errorMsg := ""
		defer func() {
			if err := recover(); err != nil {
				errorMsg = fmt.Sprintf("%+v", err)
			}
			log(ctx, start, time.Now(), info.FullMethod, errorMsg)
		}()
		err := handler(srv, ss)
		if err != nil {
			errorMsg = err.Error()
		}
		return err
	}
}

func TracingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.Pairs()
		}
		spanTracing := tracing.FromContextOrNew(ctx)
		md[strings.ToLower(tracing.TracingContextKey)] = []string{tracing.Encoding(spanTracing)}
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}
