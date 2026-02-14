package limiter

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const luaTokenBucket = `
-- 1. Get current time from Redis source of truth
local redis_time = redis.call('TIME')
local now = tonumber(redis_time[1]) 

local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])

-- 2. Fetch current state
local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
local tokens = tonumber(bucket[1]) or capacity
local last_time = tonumber(bucket[2]) or now

-- 3. Calculate refill based on Redis server time
local delta = math.max(0, now - last_time)
tokens = math.min(capacity, tokens + (delta * rate))

-- 4. Determine if request is allowed
local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

-- 5. Save state and set expiry (auto-cleanup idle buckets)
redis.call('HMSET', key, 'tokens', tokens, 'last_time', now)
redis.call('EXPIRE', key, 60)

-- 6. Calculate reset: Unix timestamp when bucket will be full
local missing = capacity - tokens
local reset_at = now + math.ceil(missing / rate)

return {allowed, math.floor(tokens), reset_at}
`

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func (r *RedisStore) Take(ctx context.Context, key string, amount int, rate float64, cap float64) (Result, error) {
	// We no longer pass time.Now() from Go!
	// ARGV[1] = rate, ARGV[2] = capacity
	res, err := r.rdb.Eval(ctx, luaTokenBucket, []string{key}, rate, cap).Result()
	if err != nil {
		// Fail-open: If Redis is down, we allow the traffic to prevent an outage
		return Result{Allowed: true}, err
	}

	vals := res.([]interface{})
	return Result{
		Allowed:   vals[0].(int64) == 1,
		Remaining: int(vals[1].(int64)),
		Reset:     vals[2].(int64),
	}, nil
}
