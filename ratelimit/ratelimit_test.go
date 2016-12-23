package lib

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RateLimitSuite struct {
	suite.Suite
	redisPool *redis.Pool
}

func (suite *RateLimitSuite) SetupSuite() {
	config.Load()
	suite.redisPool = repo.LoadRedisPool()
}

func (suite *RateLimitSuite) TearDownSuite() {
	suite.redisPool.Close()
}

func (suite *RateLimitSuite) SetupTest() {
	tu.CleanRedis(suite.redisPool)
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (suite *RateLimitSuite) TestWhenKeyDoesNotExists() {
	config.Load()
	redisPool := repo.LoadRedisPool()

	rl := RateLimit{RedisPool: redisPool}

	err := rl.Run("foo")
	require.NoError(suite.T(), err, "should not fail to rate limit")

	conn := redisPool.Get()
	result, err := redis.Int(conn.Do("GET", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")

	expiry, err := redis.Int(conn.Do("TTL", "foo"))
	require.NoError(suite.T(), err, "failed to get value from redis")

	require.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 1, result)
	assert.Equal(suite.T(), config.RateLimit().CustomerLoginWindowInMinutes()*60, expiry)
}

func (suite *RateLimitSuite) TestKeyExistsAndAttemptIsValid() {
	config.Load()
	redisPool := repo.LoadRedisPool()

	rl := RateLimit{RedisPool: redisPool}

	var err error
	allowedLoginAttempts := config.RateLimit().CustomerLoginAttempts()

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
	config.Load()
	redisPool := repo.LoadRedisPool()

	rl := RateLimit{RedisPool: redisPool}

	var err error
	allowedLoginAttempts := config.RateLimit().CustomerLoginAttempts()

	for i := 0; i < allowedLoginAttempts; i++ {
		err = rl.Run("foo")
	}
	require.NoError(suite.T(), err, "should not fail to rate limit")

	err = rl.Run("foo")
	require.Error(suite.T(), err, "should have blocked")

	assert.Equal(suite.T(), ErrBlocked, err)
}

func (suite *RateLimitSuite) TestKeyExistsAndHasWayExceededAttemptThreshold() {
	config.Load()
	redisPool := repo.LoadRedisPool()

	rl := RateLimit{RedisPool: redisPool}

	var err error
	allowedLoginAttempts := config.RateLimit().CustomerLoginAttempts()

	for i := 0; i < allowedLoginAttempts+1; i++ {
		err = rl.Run("foo")
	}

	err = rl.Run("foo")
	require.Error(suite.T(), err, "should have blocked")

	assert.Equal(suite.T(), ErrBlocked, err)
}
