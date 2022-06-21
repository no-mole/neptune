package output

import (
	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/enum"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Json(ctx *gin.Context, enum enum.ErrorNum, data interface{}) {
	ctx.JSON(enum.GetHttpCode(), &Result{
		Code: enum.GetCode(),
		Msg:  enum.GetMsg(),
		Data: data,
	})
}

func JsonNoTag(ctx *gin.Context, enum enum.ErrorNum, data interface{}) {
	ctx.Render(enum.GetHttpCode(), nJson{Data: &Result{
		Code: enum.GetCode(),
		Msg:  enum.GetMsg(),
		Data: data,
	}})
}

func File(ctx *gin.Context, filePath string) {
	ctx.File(filePath)
}
