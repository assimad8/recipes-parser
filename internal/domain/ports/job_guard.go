package ports

import "context"

type JobGuard interface {
	Acquire(ctx context.Context, key string) (string, bool, error)
	Release(ctx context.Context, key, token string) error
}