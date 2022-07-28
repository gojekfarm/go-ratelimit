package ratelimit

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RateLimitSuite struct {
	suite.Suite
	redisPool   *redis.Pool
	redisConfig *RateLimitConfig
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
	suite.redisConfig = NewRateLimitConfig(3, 15, 60)
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
	require.NoError(suite.T(), err, "should not have thrown an error")

	conn := redisPool.Get()
	result, err := redis.Int(conn.Do("GET", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")

	assert.Equal(suite.T(), allowedLoginAttempts+1, result)

	expiry, err := redis.Int(conn.Do("TTL", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")

	assert.Equal(suite.T(), redisConfig.CooldownInSeconds, expiry)
}

func (suite *RateLimitSuite) TestRateLimitNotExceeded() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	rl := RateLimit{redisPool: redisPool, config: redisConfig}

	key := "test_key"

	assert.False(suite.T(), rl.RateLimitExceeded(key))
}

func (suite *RateLimitSuite) TestRateLimitExceeded() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	var err error
	allowedLoginAttempts := redisConfig.Attempts

	key := "test_key"

	rl := RateLimit{redisPool: redisPool, config: redisConfig}

	for i := 0; i <= allowedLoginAttempts; i++ {
		err = rl.Run(key)
	}
	require.NoError(suite.T(), err, "should not fail to rate limit")

	assert.True(suite.T(), rl.RateLimitExceeded(key))
}

func (suite *RateLimitSuite) TestResetRateLimit() {
	redisPool := suite.redisPool
	redisConfig := suite.redisConfig

	allowedLoginAttempts := redisConfig.Attempts

	var err error
	rl := RateLimit{redisPool: redisPool, config: redisConfig}

	key := "test_key"

	assert.False(suite.T(), rl.RateLimitExceeded(key))

	for i := 0; i <= allowedLoginAttempts; i++ {
		err = rl.Run(key)
	}
	require.NoError(suite.T(), err, "should not fail to rate limit")

	require.True(suite.T(), rl.RateLimitExceeded(key))

	rl.Reset(key)

	assert.False(suite.T(), rl.RateLimitExceeded(key))
}
