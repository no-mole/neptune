package bar

import (
	"context"
	"fmt"
	"github.com/no-mole/neptune/protos/bar"
)

type Service struct {
	bar.UnimplementedServiceServer
}

func (s Service) SayHelly(ctx context.Context, req *bar.SayHelloRequest) (*bar.SayHelloResponse, error) {
	return &bar.SayHelloResponse{Reply: fmt.Sprintf("reply %s", req.GetSay())}, nil
}

var _ bar.ServiceServer = &Service{}
