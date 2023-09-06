package boot

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/server"
)

func HttpServer(_ context.Context) application.Plugin {
	gin.SetMode(gin.ReleaseMode)
	ginEngine := gin.New()
	return server.NewHttpServerPlugin(ginEngine)
}
