package config

type RateLimitConfig struct {
	Attempts          int
	WindowInSeconds   int
	CooldownInSeconds int
}

func NewRateLimitConfig(attempts, windowInSeconds, cooldownInSeconds int) *RateLimitConfig {
	return &RateLimitConfig{
		Attempts:          attempts,
		WindowInSeconds:   windowInSeconds,
		CooldownInSeconds: cooldownInSeconds,
	}
}
