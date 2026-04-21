package utils

import (
	"context"
	"fmt"
	"sync"

	"load-balancer/config"
	"load-balancer/models"

	"github.com/go-redis/redis/v8"
)

var tokenBucketMutex sync.Mutex

// Define the Lua script
var slidingWindowScript = redis.NewScript(`
local curr_key = KEYS[1]
local current_time = redis.call('TIME')
local current_seconds = tonumber(current_time[1])
local window_size = tonumber(ARGV[1])
local max_requests = tonumber(ARGV[2])

-- Use precise timestamp combining seconds and microseconds
local precise_time = current_seconds .. "." .. current_time[2]
local trim_time = current_seconds - window_size

-- Clean old entries and count current window
redis.call('ZREMRANGEBYSCORE', curr_key, '-inf', trim_time)
local request_count = redis.call('ZCARD', curr_key)

-- Check and update
if request_count < max_requests then
    redis.call('ZADD', curr_key, current_seconds, precise_time)
    -- Set expiration to window_size + 1 to account for edge cases
    redis.call('EXPIRE', curr_key, window_size + 1)
    return {0, max_requests - request_count - 1}
end
return {1, 0}
`)

// AllowRequestSlidingWindow implements a sliding window rate limiter
func AllowRequestSlidingWindow(redisClient *redis.Client, port int, windowSize int, maxRequests int) bool {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:%d", port)

	// Execute the Redis Lua script
	res, err := slidingWindowScript.Run(ctx, redisClient, []string{redisKey},
		windowSize,  // ARGV[1]: Sliding window size in seconds
		maxRequests, // ARGV[2]: Maximum allowed requests
	).Result()

	if err != nil {
		fmt.Printf("Redis Lua script error for port %d: %v\n", port, err)
		return true // Fallback to allow requests if Redis fails
	}

	// Parse the result from the Lua script
	result, ok := res.([]interface{})
	if !ok || len(result) < 2 {
		fmt.Printf("Unexpected Redis script result for port %d: %v\n", port, res)
		return true
	}

	allowed := result[0].(int64)
	remaining := result[1].(int64)

	if allowed == 0 {
		fmt.Printf("Request allowed for port %d. Remaining requests: %d\n", port, remaining)
		return true
	}

	fmt.Printf("Rate limit exceeded for port %d. Remaining requests: %d\n", port, remaining)
	return false
}

// Helper to fetch the latest VirtualService by port
func getVirtualServiceByPort(port int) *models.VirtualService {
	for _, vs := range config.VirtualServices {
		if vs.Port == port {
			return vs
		}
	}
	return nil
}
func GCD(serverList []*models.Server) int {
	gcd := serverList[0].Weight
	for _, server := range serverList {
		gcd = gcdHelper(gcd, server.Weight)
	}
	return gcd
}
func gcdHelper(a, b int) int {
	if b == 0 {
		return a
	}
	return gcdHelper(b, a%b)
}
func MaxWeight(serverList []*models.Server) int {
	maxWeight := 0
	for _, server := range serverList {
		if server.Weight > maxWeight {
			maxWeight = server.Weight
		}
	}
	return maxWeight
}
