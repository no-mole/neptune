package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/enum"
	"github.com/no-mole/neptune/output"
)

func GinRecover() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.Header("Gin-Server-Error", fmt.Sprintf("%v", err))
				output.Json(ctx, &enum.ErrorNumEntry{
					Code:     500,
					Msg:      fmt.Sprintf("%v", err),
					HttpCode: 500,
				}, nil)
			}
		}()
		ctx.Next()
	}
}
