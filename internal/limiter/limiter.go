package limiter

import "context"

type Result struct {
	Allowed   bool
	Remaining int
	Reset     int64
}
type LimiterStore interface {
	Take(ctx context.Context, key string, amount int) (Result, error)
}
