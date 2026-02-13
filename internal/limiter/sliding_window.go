package limiter

import (
	"sync"

	"time"
)

type UserLimiter struct {
	mu       sync.RWMutex
	requests []time.Time
}

type Guard struct {
	mu    sync.Mutex
	users map[string]*UserLimiter
	limit int
}

func NewGuard(limit int) *Guard {
	return &Guard{
		users: make(map[string]*UserLimiter),
		limit: limit,
	}
}

func (g *Guard) Allow(userId string) bool {
	g.mu.Lock()
	user, exists := g.users[userId]
	if !exists {
		user = &UserLimiter{}
		g.users[userId] = user
	}
	g.mu.Unlock()

	return user.check(g.limit)
}

func (u *UserLimiter) check(limit int) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	var validRequests []time.Time
	for _, t := range u.requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	u.requests = validRequests

	if len(u.requests) < limit {
		u.requests = append(u.requests, now)
		return true
	}
	return false
}
