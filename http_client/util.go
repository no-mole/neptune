package http_client

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const (
	// _StatusReadRespErr read resp body err, should re-call doHTTP again.
	_StatusReadRespErr = -204
	// _StatusDoReqErr do req err, should re-call doHTTP again.
	_StatusDoReqErr = -500
)

func doHTTP(ctx context.Context, method, url string, header map[string][]string, payload []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, -1, errors.Wrapf(err, "new request %s %s err", method, url)
	}

	for k, v := range header {
		req.Header.Set(k, v[0]) // multi values not supported in http option
	}

	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = errors.Wrapf(err, "do request %s %s err", method, url)
		return nil, _StatusDoReqErr, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrapf(err, "read resp body from %s %s err", method, url)
		return nil, _StatusReadRespErr, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, newReplyErr(
			resp.StatusCode,
			body,
			errors.Errorf("do %s %s return code: %d message: %s", method, url, resp.StatusCode, string(body)),
		)
	}

	return body, http.StatusOK, nil
}

// AddFormValuesIntoURL append url.Values into url string
func AddFormValuesIntoURL(rawURL string, form url.Values) (string, error) {
	if rawURL == "" {
		return "", errors.New("rawURL required")
	}
	if len(form) == 0 {
		return "", errors.New("form required")
	}

	target, err := url.Parse(rawURL)
	if err != nil {
		return "", errors.Wrapf(err, "parse rawURL `%s` err", rawURL)
	}

	urlValues := target.Query()
	for key, values := range form {
		for _, value := range values {
			urlValues.Add(key, value)
		}
	}

	target.RawQuery = urlValues.Encode()
	return target.String(), nil
}
