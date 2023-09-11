package grpc_service

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
)

type RegisterInterface interface {
	Register(ctx context.Context, service Metadata, endpoint string) error
	Unregister(ctx context.Context, service Metadata, endpoint string) error
	Close() error
}

var (
	instance                   RegisterInterface = &nop{}
	errorDefaultRegisterNotSet                   = errors.New("default register not set")
)

func SetDefaultRegister(r RegisterInterface) {
	instance = r
}

// Register 使用默认注册器注册服务
func Register(ctx context.Context, endpoint string, mds ...Metadata) error {
	if instance == nil {
		return errorDefaultRegisterNotSet
	}
	for _, md := range mds {
		err := instance.Register(ctx, md, endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// Unregister 使用默认注册器注销服务
func Unregister(ctx context.Context, endpoint string, mds ...Metadata) error {
	if instance == nil {
		return errorDefaultRegisterNotSet
	}
	for _, md := range mds {
		err := instance.Unregister(ctx, md, endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭默认注册器
func Close() error {
	return instance.Close()
}

type ServiceInfo struct {
	Item      Metadata
	Endpoints map[string]struct{}
}

func (e *ServiceInfo) Put(endpoint string) {
	if e.Endpoints == nil {
		e.Endpoints = map[string]struct{}{endpoint: {}}
	} else {
		e.Endpoints[endpoint] = struct{}{}
	}
}

func (e *ServiceInfo) Del(endpoint string) {
	delete(e.Endpoints, endpoint)
}

func (e *ServiceInfo) IsEmpty() bool {
	return len(e.Endpoints) == 0
}

type Metadata interface {
	ServiceDesc() *grpc.ServiceDesc
	Version() string
	UniqueKey() string
}

var _ Metadata = &metadata{}

func NewServiceMetadata(sd *grpc.ServiceDesc, svcVersion string) Metadata {
	return &metadata{
		sd:        sd,
		version:   svcVersion,
		uniqueKey: fmt.Sprintf("/%s/%s/%s", sd.Metadata, sd.ServiceName, svcVersion),
	}
}

type metadata struct {
	sd        *grpc.ServiceDesc
	version   string
	uniqueKey string
}

func (m *metadata) ServiceDesc() *grpc.ServiceDesc {
	return m.sd
}

func (m *metadata) Version() string {
	return m.version
}

func (m *metadata) UniqueKey() string {
	return m.uniqueKey
}
