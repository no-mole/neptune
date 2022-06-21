package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/tracing"
)

func StartTracing(ctx *gin.Context) {
	start := time.Now()
	ctxTracingValue := ctx.GetHeader(tracing.TracingContextKey)
	if ctxTracingValue != "" {
		if t, _ := tracing.Decoding(ctxTracingValue); t != nil {
			//set parent tracings
			ctx.Set(tracing.TracingContextKey, t)
		}
	}
	trac := tracing.FromContextOrNew(tracing.Start(ctx, ctx.HandlerName()))
	ctx.Set(tracing.TracingContextKey, trac)
	ctx.Header("Tracing-Id", trac.Id)
	defer func() {
		logger.Info(ctx, "tracing",
			logger.WithField("start_time", start.Format(time.RFC3339)),
			logger.WithField("end_time", time.Now().Format(time.RFC3339)),
			logger.WithField("duration", time.Since(start).Milliseconds()),
			logger.WithField("host", ctx.Request.Host),
			logger.WithField("url", ctx.Request.URL.String()),
			logger.WithField("method", ctx.Request.Method),
			logger.WithField("caller", ctx.HandlerName()),
			logger.WithField("http_code", ctx.Writer.Status()),
		)
	}()
	ctx.Next()
}
