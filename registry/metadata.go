package registry

import (
	"fmt"

	"github.com/no-mole/neptune/config"
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
	//todo remove GlobalConfig
	return fmt.Sprintf("/%s%s/%s/%s:%d", m.Namespace, m.Version, m.ServiceName, config.GlobalConfig.IP, config.GlobalConfig.GrpcPort)
}

type GrpcMeta interface {
	Metadata() *Metadata
	GenKey() string
}
