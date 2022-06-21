package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/no-mole/neptune/cache"
	"github.com/no-mole/neptune/enum"
	"github.com/no-mole/neptune/json"
	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/output"
	"github.com/no-mole/neptune/tracing"
	"golang.org/x/sync/singleflight"
)

var (
	CacheStatusKey    = "Cache-Status"
	CacheStatusHit    = "Hit"
	CacheStatusSource = "Source"
	CacheStatusShared = "Shared "
	refreshKey        = "__refresh"
)

type cacheBody struct {
	Status int         `json:"status"`
	Header http.Header `json:"header"`
	Data   []byte      `json:"data"`
}

type cacheWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Cache(store cache.Cache, expire time.Duration) gin.HandlerFunc {
	sf := singleflight.Group{}

	return func(ctx *gin.Context) {
		// only allow get request
		if ctx.Request.Method != http.MethodGet {
			return
		}

		//generate cache key
		cacheKey := genKey(ctx.Request.URL.Path, ctx.Request.URL.Query())

		// __refresh 强制缓存刷新
		_, refresh := ctx.Get(refreshKey)

		//指定__refresh不走缓存
		if !refresh {
			cachedResponse, err := store.Get(ctx, cacheKey)
			//走缓存不回源，先Abort
			if err == nil && len(cachedResponse.([]byte)) > 0 {
				ctx.Header(CacheStatusKey, CacheStatusHit)
				ctx.Abort()
				body := &cacheBody{}
				err = json.Unmarshal(cachedResponse.([]byte), body)
				if err != nil {
					outputErr(ctx, &enum.ErrorNumEntry{
						Code:     2000,
						Msg:      "cache err",
						HttpCode: 200,
					}, err)
					return
				}
				err = writeData(ctx, body)
				if err != nil {
					outputErr(ctx, &enum.ErrorNumEntry{
						Code:     2000,
						Msg:      "cache err",
						HttpCode: 200,
					}, err)
				}
				return
			}
		}
		writer := &cacheWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = writer
		var body *cacheBody

		resp, _, shared := sf.Do(cacheKey, func() (interface{}, error) {
			t := time.AfterFunc(200*time.Millisecond, func() {
				sf.Forget(cacheKey)
			})
			defer t.Stop()
			//先写header后写body（在 next 中）
			ctx.Header(CacheStatusKey, CacheStatusSource)
			ctx.Next()
			body = &cacheBody{
				Status: writer.Status(),
				Header: writer.Header(),
				Data:   writer.body.Bytes(),
			}
			return body, nil
		})
		ctx.Writer = writer.ResponseWriter
		//只缓存200的结果
		if !shared {
			if !ctx.IsAborted() && writer.Status() == 200 {
				go func() {
					//使用新context避免cancel
					data, _ := json.Marshal(body)
					_ = store.SetEx(tracing.WithContext(context.Background(), tracing.FromContextOrNew(ctx)), cacheKey, data, expire)
				}()
			}
			//body 已经写入，无需再写，直接返回
			return
		}
		//shared 的需要写入数据并且Abort
		ctx.Abort()
		ctx.Header(CacheStatusKey, CacheStatusShared)
		err := writeData(ctx, resp.(*cacheBody))
		if err != nil {
			outputErr(ctx, &enum.ErrorNumEntry{
				Code:     2001,
				Msg:      "cache err",
				HttpCode: 200,
			}, err)
			return
		}
	}
}

func genKey(prefix string, m url.Values) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf := bytes.NewBuffer([]byte{})
	for _, key := range keys {
		if key == refreshKey {
			continue
		}
		buf.WriteString(key)
		buf.WriteByte('=')
		sort.Strings(m[key])
		buf.WriteString(strings.Join(m[key], ","))
		buf.WriteByte('&')
	}
	hash := md5.Sum(buf.Bytes())
	return prefix + "_" + hex.EncodeToString(hash[:])
}

func outputErr(ctx *gin.Context, e enum.ErrorNum, err error) {
	output.Json(ctx, e, err.Error())
	logger.Error(ctx, "app", err)
	ctx.Abort()
}

func writeData(ctx *gin.Context, req *cacheBody) error {
	for k, v := range req.Header {
		for _, vv := range v {
			if ctx.Writer.Header().Get(k) != "" {
				//已经存在的不覆盖
				continue
			}
			ctx.Writer.Header().Add(k, vv)
		}
	}
	ctx.Writer.WriteHeader(req.Status)
	_, err := ctx.Writer.Write(req.Data)
	return err
}
