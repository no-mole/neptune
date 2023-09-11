package grpc_service

import (
	"net/url"
	"testing"
)

func TestTarget(t *testing.T) {
	uri := "neptune-etcd:///zeus/zeus.proto/zeus.ZeusService/v1"
	u, err := url.Parse(uri)
	if err != nil {
		t.Error(err)
	}
	t.Log(u)
}
