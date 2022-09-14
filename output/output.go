package output

import (
	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/enum"
	"net/http"
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

type yamlRender struct {
	Data []byte
}

func (r yamlRender) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	_, err := w.Write(r.Data)
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (r yamlRender) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{"application/x-yaml; charset=utf-8"})
}

func Yaml(ctx *gin.Context, data interface{}) {
	if bytes, ok := data.([]byte); !ok {
		ctx.YAML(http.StatusOK, data)
	} else {
		ctx.Render(http.StatusOK, yamlRender{Data: bytes})
	}
}
