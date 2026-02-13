package limiter

import (
	"context"
	_ "embed" // Used to load the Lua script file if you prefer
	"time"

	"github.com/redis/go-redis/v9"
)

// luaTokenBucket is the script that runs atomically inside Redis.
// It calculates refills and subtracts tokens in one step.
const luaTokenBucket = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

-- 1. Get current state
local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
local tokens = tonumber(bucket[1]) or capacity
local last_time = tonumber(bucket[2]) or now

-- 2. Calculate refill
local delta = math.max(0, now - last_time)
local refill = delta * rate
tokens = math.min(capacity, tokens + refill)

-- 3. Check and subtract
local allowed = 0
if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
end

-- 4. Save and set expiry (1 minute)
redis.call('HMSET', key, 'tokens', tokens, 'last_time', now)
redis.call('EXPIRE', key, 60)

return allowed
`

type RedisStore struct {
	rdb  *redis.Client
	rate float64
	cap  float64
}

func NewRedisStore(rdb *redis.Client, rate float64, cap float64) *RedisStore {
	return &RedisStore{
		rdb:  rdb,
		rate: rate,
		cap:  cap,
	}
}

func (r *RedisStore) Take(ctx context.Context, key string, amount int) (bool, error) {
	// We pass the current Unix time so the script is deterministic
	now := time.Now().Unix()

	result, err := r.rdb.Eval(ctx, luaTokenBucket, []string{key}, now, r.rate, r.cap, amount).Result()
	if err != nil {
		return false, err
	}

	// Redis returns 1 for true, 0 for false
	return result.(int64) == 1, nil
}
