package ratelimit

import (
	"errors"
	"go-ratelimit/config"

	"github.com/garyburd/redigo/redis"
)

// ErrBlocked is returned when ratelimiting kicks in
var ErrBlocked = errors.New("rate limit: blocked")

// RateLimit Primary struct for ratelimiting
type RateLimit struct {
	redisPool *redis.Pool
	config    *config.RateLimitConfig
}

// NewRateLimit Construction of RateLimit using RedisPool and Config
func NewRateLimit(redisPool *redis.Pool, config *config.RateLimitConfig) *RateLimit {
	return &RateLimit{
		redisPool: redisPool,
		config:    config,
	}
}

// Run applies ratelimiting to the key provided
func (rl *RateLimit) Run(key string) error {
	conn := rl.redisPool.Get()
	defer conn.Close()

	value, err := redis.Int(conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		return err
	}

	if value == 0 {
		return initializeCounterForKey(key, conn, rl.config.WindowInSeconds)
	}

	if value < rl.config.Attempts {
		return incrementCounterForKey(key, conn)
	}

	if value == rl.config.Attempts {
		return initializeCooldownWindowForKey(key, conn, rl.config.CooldownInSeconds)
	}

	return ErrBlocked
}

func initializeCounterForKey(key string, conn redis.Conn, ttl int) error {
	if _, err := conn.Do("SET", key, 1); err != nil {
		return err
	}

	if _, err := conn.Do("EXPIRE", key, ttl); err != nil {
		return err
	}

	return nil
}

func incrementCounterForKey(key string, conn redis.Conn) error {
	if _, err := conn.Do("INCR", key); err != nil {
		return err
	}

	return nil
}

func initializeCooldownWindowForKey(key string, conn redis.Conn, ttl int) error {
	if _, err := conn.Do("INCR", key); err != nil {
		return err
	}

	if _, err := conn.Do("EXPIRE", key, ttl); err != nil {
		return err
	}

	return ErrBlocked
}
