package limiter

import (
	"context"
	"math"
	"sync"
	"time"
)

type bucketState struct {
	tokens   float64
	lastTick time.Time
}

type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*bucketState
	rate float64
	cap  float64
}

func NewMemoryStore(rate float64, cap float64) *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*bucketState),
		rate: rate,
		cap:  cap,
	}
}

func (m *MemoryStore) Take(ctx context.Context, key string, amount int) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.data[key]
	if !exists {
		state = &bucketState{
			tokens:   m.cap,
			lastTick: time.Now(),
		}

		m.data[key] = state
	}
	now := time.Now()
	delta := now.Sub(state.lastTick).Seconds()
	refill := delta * m.rate

	state.tokens = math.Min(m.cap, state.tokens+refill)
	state.lastTick = now

	if state.tokens >= float64(amount) {
		state.tokens -= float64(amount)
		return true, nil
	}
	return false, nil
}
