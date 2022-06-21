package enum

import (
	"context"
)

type ErrorNum interface {
	WithError(err error) ErrorNum
	GetError() error

	WithContext(ctx context.Context) ErrorNum
	Context() context.Context

	WithCode(code int) ErrorNum
	GetCode() int

	WithMsg(desc string) ErrorNum
	GetMsg() string

	WithHttpCode(code int) ErrorNum
	GetHttpCode() int
}

type ErrorNumEntry struct {
	Code     int             // 业务码
	Msg      string          // 美化描述
	ctx      context.Context //上下文
	err      error           //发生的错误
	HttpCode int
}

func (e ErrorNumEntry) WithError(err error) ErrorNum {
	e.err = err
	return e
}

func (e ErrorNumEntry) GetError() error {
	return e.err
}

func (e ErrorNumEntry) WithContext(ctx context.Context) ErrorNum {
	e.ctx = ctx
	return e
}

func (e ErrorNumEntry) Context() context.Context {
	if e.ctx != nil {
		return e.ctx
	}
	return context.Background()
}

func (e ErrorNumEntry) WithCode(code int) ErrorNum {
	e.Code = code
	return e
}

func (e ErrorNumEntry) GetCode() int {
	return e.Code
}

func (e ErrorNumEntry) WithMsg(desc string) ErrorNum {
	e.Msg = desc
	return e
}

func (e ErrorNumEntry) GetMsg() string {
	return e.Msg
}

func (e ErrorNumEntry) WithHttpCode(code int) ErrorNum {
	e.HttpCode = code
	return e
}

func (e ErrorNumEntry) GetHttpCode() int {
	if e.HttpCode == 0 {
		return 200
	}
	return e.HttpCode
}
