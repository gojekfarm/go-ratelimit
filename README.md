[![Build Status](https://travis-ci.org/gojek-engineering/go-ratelimit.svg?branch=master)](https://travis-ci.org/gojek-engineering/go-ratelimit)

# ratelimit
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

RateLimitConfig type for setting rate limiting params

#### func  NewRateLimitConfig

```go
func NewRateLimitConfig(attempts, windowInSeconds, cooldownInSeconds int) *RateLimitConfig
```
NewRateLimitConfig to create a new config for rate limiting

```go
var ErrBlocked = errors.New("rate limit: blocked")
```
ErrBlocked is returned when an attribute is rate limited

#### type RateLimit

```go
type RateLimit struct {
}
```

RateLimit type for ratelimiting

#### func  NewRateLimit

```go
func NewRateLimit(redisPool *redis.Pool, config *config.RateLimitConfig) *RateLimit
```
NewRateLimit func to create a new rate limiting type

#### func (*RateLimit) RateLimitExceeded

```go
func (rl *RateLimit) RateLimitExceeded(key string) bool
```
RateLimitExceeded returns state of a RateLimit for a key given

#### func (*RateLimit) Reset

```go
func (rl *RateLimit) Reset(key string) error
```
Reset func clears the key from rate limiting

#### func (*RateLimit) Run

```go
func (rl *RateLimit) Run(key string) error
```
Run initiates ratelimiting for the key given

#### type RateLimiter

```go
type RateLimiter interface {
	Run(key string) error
	RateLimitExceeded(key string) bool
	Reset(key string) error
}
```

RateLimiter interface for rate limiting key

```go
Use clusterratelimiter.NewRateLimit if you are using Redis Cluster
