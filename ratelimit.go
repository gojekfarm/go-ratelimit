package ratelimit

import (
	"errors"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// ErrBlocked is returned when an attribute is rate limited
var ErrBlocked = errors.New("rate limit: blocked")

// RateLimiter interface for rate limiting key
type RateLimiter interface {
	Run(key string) error
	RateLimitExceeded(key string) bool
	Reset(key string) error
}

// RateLimit type for ratelimiting
type RateLimit struct {
	redisPool *redis.Pool
	config    *RateLimitConfig
}

// NewRateLimit func to create a new rate limiting type
func NewRateLimit(redisPool *redis.Pool, config *RateLimitConfig) *RateLimit {
	return &RateLimit{
		redisPool: redisPool,
		config:    config,
	}
}

// Run initiates ratelimiting for the key given
func (rl *RateLimit) Run(key string) error {
	conn := rl.redisPool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("INCR", key)
	conn.Send("EXPIRE", key, rl.config.WindowInSeconds)

	_, err := conn.Do("EXEC")
	if err != nil {
		return err
	}

	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if result > rl.config.Attempts {
		_, err := conn.Do("EXPIRE", key, rl.config.CooldownInSeconds)
		if err != nil {
			return err
		}

	}

	return nil
}

// RateLimitExceeded returns state of a RateLimit for a key given
func (rl *RateLimit) RateLimitExceeded(key string) bool {
	conn := rl.redisPool.Get()
	defer conn.Close()

	value, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		return false
	}

	if value > rl.config.Attempts {
		return true
	}

	return false
}

// Reset func clears the key from rate limiting
func (rl *RateLimit) Reset(key string) error {
	conn := rl.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}

	return nil
}
