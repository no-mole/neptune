package http_client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/no-mole/neptune/logger"
	"github.com/no-mole/neptune/tracing"
	"github.com/pkg/errors"
)

const (
	// DefaultTTL the entire time cost, contains N times retry
	DefaultTTL = time.Minute
	// DefaultRetryTimes retry how many times
	DefaultRetryTimes = 3
	// DefaultRetryDelay delay some time before retry
	DefaultRetryDelay = time.Millisecond * 100
)

// TODO retry的code不一定正确，缺失或者多余待实际使用中修改。
func shouldRetry(ctx context.Context, httpCode int) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	switch httpCode {
	case
		_StatusDoReqErr,    // customize
		_StatusReadRespErr, // customize

		http.StatusRequestTimeout,
		http.StatusLocked,
		http.StatusTooEarly,
		http.StatusTooManyRequests,

		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:

		return true

	default:
		return false
	}
}

// PostForm post some form values
func PostForm(ctx context.Context, url string, form url.Values, options ...Option) (body []byte, err error) {
	if url == "" {
		return nil, errors.New("url required")
	}
	if len(form) == 0 {
		return nil, errors.New("form required")
	}

	opt := getOption()
	for _, f := range options {
		f(opt)
	}

	ctx = tracing.Start(ctx, "http_client")
	trace := tracing.FromContextOrNew(ctx)
	str := tracing.Encoding(trace)
	defer func() {
		logger.Info(ctx, "tracing", logger.WithField("url", url))
	}()
	opt.header["Tracing-Context-Key"] = []string{str}
	opt.header["Content-Type"] = []string{"application/x-www-form-urlencoded; charset=utf-8"}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	ctx, _ = context.WithTimeout(ctx, ttl)

	formValue := form.Encode()

	retryTimes := opt.retryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.retryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	var httpCode int
	for k := 0; k < retryTimes; k++ {
		body, httpCode, err = doHTTP(ctx, http.MethodPost, url, opt.header, []byte(formValue))
		if shouldRetry(ctx, httpCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

// PostJSON post some json
func PostJSON(ctx context.Context, url string, raw json.RawMessage, options ...Option) (body []byte, err error) {
	if url == "" {
		return nil, errors.New("url required")
	}
	if len(raw) == 0 {
		return nil, errors.New("raw required")
	}

	opt := getOption()
	for _, f := range options {
		f(opt)
	}

	ctx = tracing.Start(ctx, "http_client")
	trace := tracing.FromContextOrNew(ctx)
	str := tracing.Encoding(trace)
	defer func() {
		logger.Info(ctx, "tracing", logger.WithField("url", url))
	}()
	opt.header["Tracing-Context-Key"] = []string{str}
	opt.header["Content-Type"] = []string{"application/json; charset=utf-8"}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	ctx, _ = context.WithTimeout(ctx, ttl)

	retryTimes := opt.retryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.retryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	var httpCode int
	for k := 0; k < retryTimes; k++ {
		body, httpCode, err = doHTTP(ctx, http.MethodPost, url, opt.header, raw)
		if shouldRetry(ctx, httpCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

// Get get something by form values
func Get(ctx context.Context, url string, form url.Values, options ...Option) (body []byte, err error) {
	if url == "" {
		return nil, errors.New("url required")
	}

	if len(form) > 0 {
		if url, err = AddFormValuesIntoURL(url, form); err != nil {
			return
		}
	}

	opt := getOption()
	for _, f := range options {
		f(opt)
	}

	ctx = tracing.Start(ctx, "http_client")
	trace := tracing.FromContextOrNew(ctx)
	str := tracing.Encoding(trace)
	defer func() {
		logger.Info(ctx, "tracing", logger.WithField("url", url))
	}()
	opt.header["Tracing-Context-Key"] = []string{str}
	opt.header["Content-Type"] = []string{"application/x-www-form-urlencoded; charset=utf-8"}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, _ = context.WithTimeout(ctx, ttl)

	retryTimes := opt.retryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.retryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	var httpCode int
	for k := 0; k < retryTimes; k++ {
		body, httpCode, err = doHTTP(ctx, http.MethodGet, url, opt.header, nil)
		if shouldRetry(ctx, httpCode) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}
