package storage

import "context"

type LimiterStore interface {
	Take(ctx context.Context, key string, amount int) (bool, error)
}
