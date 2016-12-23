package config

// RateLimitConfig Config object for RateLimiting
type RateLimitConfig struct {
	Attempts          int
	WindowInSeconds   int
	CooldownInSeconds int
}

// NewRateLimitConfig Construct config for RateLimiting
func NewRateLimitConfig(attempts, windowInSeconds, cooldownInSeconds int) *RateLimitConfig {
	return &RateLimitConfig{
		Attempts:          attempts,
		WindowInSeconds:   windowInSeconds,
		CooldownInSeconds: cooldownInSeconds,
	}
}
