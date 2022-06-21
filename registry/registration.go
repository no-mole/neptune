package registry

import (
	"context"
)

var (
	TTL int64 = 15
)

type Registration interface {
	Register(ctx context.Context, meta GrpcMeta) (err error)
	UnRegister(ctx context.Context, meta GrpcMeta) (err error)
}

var register Registration

func SetRegister(r Registration) {
	register = r
}

func GetRegister() Registration {
	return register
}

func Registry(ctx context.Context, meta ...GrpcMeta) error {
	for _, m := range meta {
		err := register.Register(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func UnRegister(ctx context.Context, meta GrpcMeta) (err error) {
	return register.UnRegister(ctx, meta)
}
