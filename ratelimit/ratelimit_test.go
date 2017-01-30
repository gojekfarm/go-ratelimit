package ratelimit

import (
	"go-ratelimit/config"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RateLimitSuite struct {
	suite.Suite
	redisPool   *redis.Pool
	redisConfig *config.RateLimitConfig
}

func testRedisPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}

			return c, err
		},
	}
}

func (suite *RateLimitSuite) SetupSuite() {
	suite.redisConfig = config.NewRateLimitConfig(3, 15, 60)
	suite.redisPool = testRedisPool()
}

func (suite *RateLimitSuite) TearDownSuite() {
	suite.redisPool.Close()
}

func (suite *RateLimitSuite) SetupTest() {
	conn := suite.redisPool.Get()
	conn.Do("FLUSHALL")
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (suite *RateLimitSuite) TestWhenKeyDoesNotExists() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: suite.redisConfig}

	err := rl.Run("foo")
	require.NoError(suite.T(), err, "should not fail to rate limit")

	conn := redisPool.Get()
	result, err := redis.Int(conn.Do("GET", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")

	expiry, err := redis.Int(conn.Do("TTL", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")

	require.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 1, result)
	assert.Equal(suite.T(), redisConfig.WindowInSeconds, expiry)
}

func (suite *RateLimitSuite) TestKeyExistsAndAttemptIsValid() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: suite.redisConfig}

	var err error
	allowedLoginAttempts := redisConfig.Attempts

	for i := 0; i < allowedLoginAttempts; i++ {
		err = rl.Run("foo")
	}
	require.NoError(suite.T(), err, "should not fail to rate limit")

	conn := redisPool.Get()
	result, err := redis.Int(conn.Do("GET", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")
	require.NotNil(suite.T(), result)

	assert.Equal(suite.T(), allowedLoginAttempts, result)
}

func (suite *RateLimitSuite) TestKeyExistsAndJustExceededAttemptThreshold() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: suite.redisConfig}

	var err error
	allowedLoginAttempts := redisConfig.Attempts

	for i := 0; i < allowedLoginAttempts; i++ {
		err = rl.Run("foo")
	}
	require.NoError(suite.T(), err, "should not fail to rate limit")

	err = rl.Run("foo")
	require.Error(suite.T(), err, "should have blocked")

	assert.Equal(suite.T(), ErrBlocked, err)
}

func (suite *RateLimitSuite) TestKeyExistsAndHasWayExceededAttemptThreshold() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: suite.redisConfig}

	var err error
	allowedLoginAttempts := redisConfig.Attempts

	for i := 0; i < allowedLoginAttempts+1; i++ {
		err = rl.Run("foo")
	}

	err = rl.Run("foo")
	require.Error(suite.T(), err, "should have blocked")

	assert.Equal(suite.T(), ErrBlocked, err)
}

func (suite *RateLimitSuite) TestRateLimitNotExceeded() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: redisConfig}

	key := "test_key"

	assert.False(suite.T(), rl.RateLimitExceeded(key))

	initializeCounterForKey(key, redisPool.Get(), 2)
	incrementCounterForKey(key, redisPool.Get())
	incrementCounterForKey(key, redisPool.Get())

	assert.False(suite.T(), rl.RateLimitExceeded(key))
}

func (suite *RateLimitSuite) TestRateLimitExceeded() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: redisConfig}

	key := "test_key"

	initializeCounterForKey(key, redisPool.Get(), 2)
	incrementCounterForKey(key, redisPool.Get())
	incrementCounterForKey(key, redisPool.Get())
	incrementCounterForKey(key, redisPool.Get())

	assert.True(suite.T(), rl.RateLimitExceeded(key))
}

func (suite *RateLimitSuite) TestResetRateLimit() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: redisConfig}

	key := "test_key"

	assert.False(suite.T(), rl.RateLimitExceeded(key))

	initializeCounterForKey(key, redisPool.Get(), 2)

	for i := 0; i < suite.redisConfig.Attempts; i++ {
		incrementCounterForKey(key, redisPool.Get())
	}

	require.True(suite.T(), rl.RateLimitExceeded(key))

	rl.Reset(key)

	assert.False(suite.T(), rl.RateLimitExceeded(key))
}
