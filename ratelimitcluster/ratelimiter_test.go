package clusterratelimiter

import (
	"github.com/go-redis/redis"
	"github.com/go-ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func initializeRedisClusterClient(t *testing.T) *redis.ClusterClient {
	getTimeInSec := func(tm int) time.Duration { return time.Second * time.Duration(tm) }

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:       []string{"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003", "localhost:7004", "localhost:7005"},
		ReadTimeout: getTimeInSec(5000),
		IdleTimeout: getTimeInSec(10),
		DialTimeout: getTimeInSec(5000),
	})
}

func TestNewRateLimit(t *testing.T) {
	clusterClient := initializeRedisClusterClient(t)

	config := &ratelimit.RateLimitConfig{
		Attempts:          1,
		WindowInSeconds:   10,
		CooldownInSeconds: 10,
	}

	actalRateLimit := NewRateLimit(clusterClient, config)

	expectedRateLimit := &RateLimit{
		config:        config,
		clusterClient: clusterClient,
	}

	assert.Equal(t, expectedRateLimit, actalRateLimit)
}

func beforeTest(client *redis.ClusterClient, t *testing.T) {
	cmd := client.FlushAll()
	require.NoError(t, cmd.Err())
}

func TestRateLimitRun(t *testing.T) {
	err := os.Setenv("REDIS_CLUSTER_IP", "0.0.0.0")
	require.NoError(t, err)

	getErrorMessage := func(rateLimit *RateLimit, key string) string {
		err := rateLimit.Run(key)
		if err != nil {
			return err.Error()
		}
		return ""
	}

	testCases := []struct {
		name            string
		actualMessage   func() string
		expectedMessage string
	}{
		{
			name: "test run success",
			actualMessage: func() string {

				key := "login"
				clusterClient := initializeRedisClusterClient(t)
				beforeTest(clusterClient, t)

				config := &ratelimit.RateLimitConfig{
					Attempts:          1,
					WindowInSeconds:   10,
					CooldownInSeconds: 10,
				}

				rateLimit := NewRateLimit(clusterClient, config)

				return getErrorMessage(rateLimit, key)

			},
			expectedMessage: "",
		},
		{
			name: "test run failure for invalid key",
			actualMessage: func() string {

				key := "invalid_key"

				clusterClient := initializeRedisClusterClient(t)
				beforeTest(clusterClient, t)

				clusterClient.Set(key, "invalid_data", time.Minute)

				config := &ratelimit.RateLimitConfig{
					Attempts:          1,
					WindowInSeconds:   10,
					CooldownInSeconds: 10,
				}

				rateLimit := NewRateLimit(clusterClient, config)

				return getErrorMessage(rateLimit, key)
			},
			expectedMessage: "ERR value is not an integer or out of range",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedMessage, testCase.actualMessage())
		})
	}
}

func TestRateLimitReset(t *testing.T) {
	testCases := []struct {
		name          string
		actualError   func() error
		expectedError error
	}{
		{
			name: "test reset key success",
			actualError: func() error {
				key := "login"
				clusterClient := initializeRedisClusterClient(t)
				clusterClient.Set(key, 1, 10*time.Second)

				config := &ratelimit.RateLimitConfig{
					Attempts:          1,
					WindowInSeconds:   10,
					CooldownInSeconds: 10,
				}

				rateLimit := NewRateLimit(clusterClient, config)

				return rateLimit.Reset(key)
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedError, testCase.actualError())
		})
	}
}

func TestRateLimitExceed(t *testing.T) {
	testCases := []struct {
		name           string
		actualResult   func() bool
		expectedResult bool
	}{
		{
			name: "test rate limit is exceeded",
			actualResult: func() bool {
				key := "login"

				clusterClient := initializeRedisClusterClient(t)
				beforeTest(clusterClient, t)

				config := &ratelimit.RateLimitConfig{
					Attempts:          0,
					WindowInSeconds:   10,
					CooldownInSeconds: 10,
				}

				rateLimit := NewRateLimit(clusterClient, config)

				err := rateLimit.Run(key)
				require.NoError(t, err)

				return rateLimit.RateLimitExceeded(key)

			},
			expectedResult: true,
		},
		{
			name: "test rate limit is not exceeded",
			actualResult: func() bool {
				key := "login"

				clusterClient := initializeRedisClusterClient(t)
				beforeTest(clusterClient, t)

				config := &ratelimit.RateLimitConfig{
					Attempts:          1,
					WindowInSeconds:   10,
					CooldownInSeconds: 10,
				}

				rateLimit := NewRateLimit(clusterClient, config)

				return rateLimit.RateLimitExceeded(key)

			},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedResult, testCase.actualResult())
		})
	}
}
