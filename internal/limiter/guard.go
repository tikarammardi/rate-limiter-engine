package limiter

import (
	"context"
	"fmt"
	"rate-limiter-engine/internal/storage"
)

type Guard struct {
	store storage.LimiterStore
}

func NewGuard(store storage.LimiterStore) *Guard {
	return &Guard{store: store}
}

func (g *Guard) Allow(ctx context.Context, userId string) bool {

	key := fmt.Sprintf("user:%s", userId)

	allowed, err := g.store.Take(ctx, key, 1)
	if err != nil {

		return true
	}

	return allowed
}
