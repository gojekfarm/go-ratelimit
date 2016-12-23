package config

type RateLimitConfig struct {
	Attempts          int
	WindowInMinutes   int
	CooldownInMinutes int
}

func NewRateLimitConfig(attempts, windowInMinutes, cooldownInMinutes int) *RateLimitConfig {
	return &RateLimitConfig{
		Attempts:          attempts,
		WindowInMinutes:   windowInMinutes,
		CooldownInMinutes: cooldownInMinutes,
	}
}
