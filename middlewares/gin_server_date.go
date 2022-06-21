package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func ServerDate(ctx *gin.Context) {
	ctx.Header("Server-Date", time.Now().Format("2006-01-02 15:04:05"))
}
