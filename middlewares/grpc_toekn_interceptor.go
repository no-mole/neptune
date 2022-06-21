package middleware

//
//// TokenServerInterceptor 拦截器
//func TokenServerInterceptor() grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
//		ctx = startTracingFromMetadata(ctx, info.FullMethod)
//		start := time.Now()
//		errorMsg := ""
//		defer func() {
//			err := recover()
//			if err != nil {
//				errorMsg = fmt.Sprintf("%+v", err)
//			}
//			log(ctx, start, time.Now(), info.FullMethod, errorMsg)
//		}()
//		resp, err = handler(ctx, req)
//		if err != nil {
//			errorMsg = err.Error()
//		}
//		return resp, err
//	}
//}
//
//func TokenClientInterceptor() grpc.UnaryClientInterceptor {
//	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
//		md, ok := metadata.FromOutgoingContext(ctx)
//		if !ok {
//			md = metadata.Pairs()
//		}
//		ctx.Value()
//		spanTracing := tracing.FromContextOrNew(ctx)
//		md[strings.ToLower()] = []string{tracing.Encoding(spanTracing)}
//		ctx = metadata.NewOutgoingContext(ctx, md)
//		return invoker(ctx, method, req, resp, cc, opts...)
//	}
//}
