package bar

import "github.com/no-mole/neptune/grpc_service"

var Metadata = grpc_service.NewServiceMetadata(&Service_ServiceDesc, "neptune", "v1")
