package clusterratelimiter

import (
	"github.com/go-redis/redis"
	"time"
	"github.com/go-ratelimit"
)

// RateLimit type for ratelimiting
type RateLimit struct {
	clusterClient *redis.ClusterClient
	config        *ratelimit.RateLimitConfig
}

// NewRateLimit func to create a new rate limiting type
func NewRateLimit(clusterClient *redis.ClusterClient, config *ratelimit.RateLimitConfig) *RateLimit {
	return &RateLimit{
		clusterClient: clusterClient,
		config:        config,
	}
}

// Run initiates ratelimiting for the key given
func (rl *RateLimit) Run(key string) error {
	pipeline := rl.clusterClient.TxPipeline()
	pipeline.Incr(key)
	pipeline.Expire(key, time.Second*time.Duration(rl.config.WindowInSeconds))

	_, err := pipeline.Exec()
	if err != nil {
		return err
	}

	defer func() {
		pipeline.Close()
	}()

	cmd := rl.clusterClient.Get(key)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	val, err := cmd.Int()
	if err != nil {
		return err
	}

	if val > rl.config.Attempts {
		if cmd := rl.clusterClient.Expire(key, time.Second*time.Duration(rl.config.CooldownInSeconds)); cmd.Err() != nil {
			return cmd.Err()
		}
	}

	return nil
}

// Reset func clears the key from rate limiting
func (rl *RateLimit) Reset(key string) error {
	if cmdErr := rl.clusterClient.Del(key); cmdErr.Err() != nil {
		return cmdErr.Err()
	}
	return nil
}

// RateLimitExceeded returns state of a RateLimit for a key given
func (rl *RateLimit) RateLimitExceeded(key string) bool {
	cmd := rl.clusterClient.Get(key)
	if cmd.Err() != nil {
		return false
	}

	value, err := cmd.Int()
	if err != nil {
		return false
	}

	if value > rl.config.Attempts {
		return true
	}

	return false
}
