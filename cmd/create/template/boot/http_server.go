package boot

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/application"
	"github.com/no-mole/neptune/server"
	"net/http"
)

func HttpServer(_ context.Context) application.Plugin {
	handleFn := func(ctx context.Context) http.Handler {
		gin.SetMode(gin.ReleaseMode)
		ginEngine := gin.New()
		return ginEngine
	}
	return server.NewHttpServerPlugin(handleFn)
}
