package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/logger"
)

func AccessLog(ctx *gin.Context) {
	start := time.Now()
	defer func() {
		end := time.Now()
		logger.Info(ctx, "gin",
			logger.WithField("start_time", start.Format(time.RFC3339)),
			logger.WithField("end_time", end.Format(time.RFC3339)),
			logger.WithField("duration", time.Since(start).Milliseconds()),
			logger.WithField("host", ctx.Request.Host),
			logger.WithField("url", ctx.Request.URL.String()),
			logger.WithField("method", ctx.Request.Method),
			logger.WithField("code", ctx.Writer.Status()),
			logger.WithField("caller", ctx.HandlerName()),
		)
	}()
	ctx.Next()
}
