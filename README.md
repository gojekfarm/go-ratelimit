## Go RateLimit
### A utility to perform rate limiting in golang using redis backend

--
    import "go-ratelimit/ratelimit"


## Usage

#### type RateLimitConfig

```go
type RateLimitConfig struct {
	Attempts          int
	WindowInSeconds   int
	CooldownInSeconds int
}
```

RateLimitConfig Config object for RateLimiting

#### func NewRateLimitConfig

```go
func NewRateLimitConfig(attempts, windowInSeconds, cooldownInSeconds int) *RateLimitConfig
```
NewRateLimitConfig Construct config for RateLimiting

```go
var ErrBlocked = errors.New("rate limit: blocked")
```
ErrBlocked is returned when ratelimiting kicks in

#### type RateLimit

```go
type RateLimit struct {
}
```

RateLimit Primary struct for ratelimiting

#### func  NewRateLimit

```go
func NewRateLimit(redisPool *redis.Pool, config *config.RateLimitConfig) *RateLimit
```
NewRateLimit Construction of RateLimit using RedisPool and Config

#### func (*RateLimit) RateLimitExceeded

```go
func (rl *RateLimit) RateLimitExceeded(key string) bool
```
RateLimitExceeded returns state of RateLimit for a key provided

#### func (*RateLimit) Run

```go
func (rl *RateLimit) Run(key string) error
```
Run applies ratelimiting to the key provided

#### type RateLimiter

```go
type RateLimiter interface {
	Run(key string) error
	RateLimitExceeded(key string) bool
}
```

RateLimiter interface which every RateLimit needs to implement
