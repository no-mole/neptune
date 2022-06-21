package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CompareScopesFunc func(ctx *gin.Context, scopes []string) (bool, error)

type CancelFunc func(ctx *gin.Context, err error)

var (
	ErrorTokenNotAvailable   = errors.New("token not available")
	ErrorCompareScopesFailed = errors.New("compare scopes failed")
	ErrorScopesNotCompared   = errors.New("permission denied")
)

func IntrospectionCheck(compareScopesFunc CompareScopesFunc, endpoint, clientKey, clientSecret string, cancelFunc CancelFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := getToken(ctx)
		resp, err := postIntrospection(endpoint, token, clientKey, clientSecret)
		if err != nil {
			cancelFunc(ctx, err)
			ctx.Abort()
			return
		}
		if resp.Active == false {
			cancelFunc(ctx, ErrorTokenNotAvailable)
			ctx.Abort()
			return
		}
		ok, err := compareScopesFunc(ctx, resp.Scopes)
		if err != nil {
			cancelFunc(ctx, ErrorCompareScopesFailed)
			ctx.Abort()
			return
		}
		if !ok {
			cancelFunc(ctx, ErrorScopesNotCompared)
			ctx.Abort()
			return
		}
	}
}

const tokenHeader = "Authorization"

func getToken(ctx *gin.Context) string {
	token := ctx.GetHeader(tokenHeader)
	if strings.HasPrefix(token, "Bearer") {
		token = strings.TrimSpace(strings.TrimLeft(token, "Bearer"))
	}
	return token
}

func postIntrospection(url, token, clientKey, clientSecret string) (*IntrospectionResp, error) {

	req, err := http.NewRequest(http.MethodPost, url, genPostBody(&IntrospectionReq{Token: token}))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(clientKey, clientSecret)

	p := &IntrospectionResp{}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return p, err
	}

	if resp.StatusCode != http.StatusOK {
		return p, errors.New("post introspection failed")
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return p, err
	}

	err = json.Unmarshal(data, p)
	return p, err
}

func genPostBody(req *IntrospectionReq) io.Reader {
	data, _ := json.Marshal(req)
	return bytes.NewBuffer(data)
}

type IntrospectionReq struct {
	Token         string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	TokenTypeHint string `protobuf:"bytes,2,opt,name=token_type_hint,json=tokenTypeHint,proto3" json:"token_type_hint,omitempty"`
	ClientKey     string `protobuf:"bytes,3,opt,name=client_key,json=clientKey,proto3" json:"client_key,omitempty"`
	ClientSecret  string `protobuf:"bytes,4,opt,name=client_secret,json=clientSecret,proto3" json:"client_secret,omitempty"`
}

type IntrospectionResp struct {
	Active    bool     `protobuf:"varint,1,opt,name=active,proto3" json:"active,omitempty"`
	Scopes    []string `protobuf:"bytes,2,rep,name=scopes,proto3" json:"scopes,omitempty"`
	ClientId  string   `protobuf:"bytes,3,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	Username  string   `protobuf:"bytes,4,opt,name=username,proto3" json:"username,omitempty"`
	TokenType string   `protobuf:"bytes,5,opt,name=token_type,json=tokenType,proto3" json:"token_type,omitempty"`
	Exp       int64    `protobuf:"varint,6,opt,name=exp,proto3" json:"exp,omitempty"`
	Iat       int64    `protobuf:"varint,7,opt,name=iat,proto3" json:"iat,omitempty"`
	Nbf       int64    `protobuf:"varint,8,opt,name=nbf,proto3" json:"nbf,omitempty"`
	Sub       string   `protobuf:"bytes,9,opt,name=sub,proto3" json:"sub,omitempty"`
	Aud       string   `protobuf:"bytes,10,opt,name=aud,proto3" json:"aud,omitempty"`
	Iss       string   `protobuf:"bytes,11,opt,name=iss,proto3" json:"iss,omitempty"`
	Jti       string   `protobuf:"bytes,12,opt,name=jti,proto3" json:"jti,omitempty"`
}
