package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/enum"
	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/output"
)

func Recover(ctx *gin.Context) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			logger.Error(ctx, "gin", errors.New(fmt.Sprintf("%v", err)),
				logger.WithField("errType", "recover"),
				logger.WithField("end_time", time.Now().Format(time.RFC3339)),
				logger.WithField("duration", time.Since(start).Milliseconds()),
				logger.WithField("host", ctx.Request.Host),
				logger.WithField("url", ctx.Request.URL.String()),
				logger.WithField("method", ctx.Request.Method),
				logger.WithField("caller", ctx.HandlerName()),
			)
			output.Json(ctx, &enum.ErrorNumEntry{
				Code:     500,
				Msg:      fmt.Sprintf("%v", err),
				HttpCode: 500,
			}, nil)
		}
	}()
	ctx.Next()
}
