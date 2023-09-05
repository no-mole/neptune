package grpc_service

import "context"

var _ RegisterInterface = &nop{}

type nop struct{}

func (n *nop) Register(_ context.Context, _ Metadata, _ string) error {
	return nil
}

func (n *nop) Unregister(_ context.Context, _ Metadata, _ string) error {
	return nil
}

func (n *nop) Close() error {
	return nil
}
