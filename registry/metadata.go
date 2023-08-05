package registry

import (
	"fmt"

	"google.golang.org/grpc"
)

type Metadata struct {
	grpc.ServiceDesc
	Version   string
	Namespace string
}

func (m *Metadata) Metadata() *Metadata {
	return m
}

func (m *Metadata) GenKey() string {
	return fmt.Sprintf("/%s-%s/%s/", m.Namespace, m.Version, m.ServiceName)
}

type GrpcMeta interface {
	Metadata() *Metadata
	GenKey() string
}
