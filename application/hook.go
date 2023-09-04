package application

import "context"

type HookFunc func(ctx context.Context) error
