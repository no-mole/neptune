package bar

import (
	"github.com/no-mole/neptune/registry"
)

func Metadata() *registry.Metadata {
	return &registry.Metadata{
		ServiceDesc: Service_ServiceDesc,
		Namespace:   "neptune",
		Version:     "v1",
	}
}
