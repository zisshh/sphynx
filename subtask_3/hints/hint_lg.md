1. First, create this Lua script for atomic rate limiting operations:
```lua
local curr_key = KEYS[1]
local current_time = redis.call('TIME')
local current_seconds = tonumber(current_time[1])
local window_size = tonumber(ARGV[1])
local max_requests = tonumber(ARGV[2])

-- Use precise timestamp
local precise_time = current_seconds .. "." .. current_time[2]
local trim_time = current_seconds - window_size

-- Clean old entries and count current window
redis.call('ZREMRANGEBYSCORE', curr_key, '-inf', trim_time)
local request_count = redis.call('ZCARD', curr_key)

-- Check and update
if request_count < max_requests then
    redis.call('ZADD', curr_key, current_seconds, precise_time)
    redis.call('EXPIRE', curr_key, window_size + 1)
    return {0, max_requests - request_count - 1}
end
return {1, 0}
```

2. Implement the rate limiting function:
```go
func AllowRequestSlidingWindow(redisClient *redis.Client, port int, windowSize int, maxRequests int) bool {
    ctx := context.Background()
    redisKey := fmt.Sprintf("rate_limit:%d", port)

    // Execute the Redis Lua script
    res, err := slidingWindowScript.Run(ctx, redisClient, []string{redisKey},
        windowSize,  // Window size in seconds
        maxRequests, // Maximum allowed requests
    ).Result()

    if err != nil {
        // Fallback to allow requests if Redis fails
        return true
    }

    result, ok := res.([]interface{})
    if !ok || len(result) < 2 {
        return true
    }

    allowed := result[0].(int64)
    return allowed == 0
}
```

3. Create the middleware:
```go
func MiddlewareSlidingWindowRateLimiting(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        portStr := mux.Vars(r)["vs_id"]
        port, err := strconv.Atoi(portStr)
        if err != nil {
            http.Error(w, "Invalid port format", http.StatusBadRequest)
            return
        }

        vs := getVirtualServiceByPort(port)
        if vs == nil {
            http.Error(w, "Virtual service not found", http.StatusNotFound)
            return
        }

        // Apply sliding window rate limiting
        if !AllowRequestSlidingWindow(redisClient, port, 60, vs.RateLimit) {
            http.Error(w, vs.Message, vs.StatusCode)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

4. Initialize Redis connection with proper configuration:
```go
func InitializeRedis() {
    RedisClient = redis.NewClient(&redis.Options{
        Addr: "127.0.0.1:6379",
        DB: 0,
        PoolSize: 10,
        MinIdleConns: 5,
    })
}
```

Key implementation details:
- Use `ZREMRANGEBYSCORE` to clean old entries
- Use `ZADD` with current timestamp for new requests
- Set expiration on keys to prevent memory leaks
- Handle Redis errors gracefully with fallback behavior
- Use proper connection pooling for performance
- Implement the operations atomically using Lua script

Remember to properly handle:
1. Redis connection failures
2. Configuration updates
3. Proper error responses
4. Logging of rate limit violations