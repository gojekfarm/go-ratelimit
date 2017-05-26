package ratelimit

// RateLimitConfig type for setting rate limiting params
type RateLimitConfig struct {
	Attempts          int
	WindowInSeconds   int
	CooldownInSeconds int
}

// NewRateLimitConfig to create a new config for rate limiting
func NewRateLimitConfig(attempts, windowInSeconds, cooldownInSeconds int) *RateLimitConfig {
	return &RateLimitConfig{
		Attempts:          attempts,
		WindowInSeconds:   windowInSeconds,
		CooldownInSeconds: cooldownInSeconds,
	}
}
