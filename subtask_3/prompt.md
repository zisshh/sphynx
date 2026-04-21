# Subtask 3: Distributed Rate Limiting

### Requirements

#### Application/Code Behavior

1. **Sliding Window Rate Limiting**:
   - Implement rate limiting using a sliding window algorithm with Redis
   - Configure rate limits at the virtual service level
   - Track requests within a 60-second window
   - Input Example:
     ```json
     {
       "port": 8001,
       "rate_limit": 15,
       "status_code": 429,
       "message": "Rate limit exceeded - Please try again later"
     }
     ```
   - Output Example:
     - When limit exceeded, return configured status code and message
     - Log rate limit violation with timestamp and client IP

2. **Redis-Based Rate Limiting**:
   - Use Redis sorted sets for tracking request timestamps
   - Implement atomic operations using Lua scripts
   - Cleanup expired entries automatically
   - Handle Redis connection failures gracefully
   - Input Example:
     - Redis key format: `rate_limit:{port}`
     - Window size: 60 seconds
   - Output Example:
     - Allow or deny request based on sliding window count
     - Return remaining quota information

3. **Distributed Coordination**:
   - Ensure consistent rate limiting across load balancer instances
   - Maintain synchronized state using Redis
   - Handle race conditions with atomic operations
   - Support multiple load balancer instances accessing same Redis keys

#### Expected Architectural Decisions

1. **Rate Limiter Design**:
   - Implement as middleware component
   - Use Redis for distributed state
   - Support dynamic configuration updates
   - Handle failures gracefully

2. **Redis Integration**:
   - Connect to Redis on startup
   - Use Lua scripts for atomic operations
   - Implement connection pooling
   - Handle reconnection scenarios

3. **Configuration Management**:
   - Store rate limits per virtual service
   - Support runtime updates
   - Maintain defaults for error cases
   - Persist configurations

#### Software Design Patterns

1. **Middleware Pattern**:
   - Rate limiting implemented as HTTP middleware
   - Chainable with other middleware components

2. **Repository Pattern**:
   - Abstract Redis operations
   - Enable testing and mocking

3. **Factory Pattern**:
   - Create rate limiters based on configuration
   - Support different rate limiting strategies

#### Advanced Language Features

1. **Concurrency**:
   - Goroutines for request handling
   - Mutex for thread-safe operations
   - Context for timeouts and cancellation

2. **Redis Integration**:
   - Connection pooling
   - Lua script execution
   - Error handling and retry logic

3. **Atomic Operations**:
   - Lua scripts for Redis commands
   - Thread-safe configuration updates
   - Distributed locking when needed

#### Error Handling

1. **Rate Limit Violations**:
   - Return configured status code
   - Include custom error message
   - Log violation details
   - Include remaining quota in response

2. **Redis Failures**:
   - Implement fallback behavior
   - Log connection errors
   - Attempt reconnection
   - Continue service with degraded rate limiting

3. **Configuration Errors**:
   - Validate rate limit values
   - Log configuration issues
   - Maintain existing configuration on error
   - Notify administrators of issues

#### Technical Specifications

1. **Required Functions**:
   ```go
   func MiddlewareSlidingWindowRateLimiting(next http.Handler) http.Handler
   func AllowRequestSlidingWindow(redisClient *redis.Client, port int, windowSize int, maxRequests int) bool
   ```

2. **Redis Requirements**:
   - Version: >= 6.0
   - Required commands:
     - ZADD
     - ZREMRANGEBYSCORE
     - ZCARD
     - DEL
     - EXPIRE
   - Lua script capabilities

3. **Configuration Requirements**:
   - Rate limits must be positive integers
   - Status codes must be valid HTTP status codes
   - Messages must not exceed 256 characters
   - Window size fixed at 60 seconds

4. **Performance Requirements**:
   - Rate limiting decision under 10ms
   - Redis operations timeout after 100ms
   - Graceful degradation on Redis failures
   - Efficient memory usage for sliding window

5. **Logging Requirements**:
   - Log all rate limit violations
   - Log configuration changes
   - Log Redis connection issues
   - Include timestamp, service port, client IP

### API Specifications

#### 1. Configure Rate Limits

**Endpoint:** POST /access/vs/rate-limits

##### Successful Request
**Input:**
```json
{
    "port": 8001,
    "rate_limit": 15,
    "status_code": 429,
    "message": "Rate limit exceeded - Please try again later."
}
```

**Response:**
- Status Code: 200 OK
- Body:
```json
{
    "status": "success",
    "message": "Rate limits updated and Redis keys reset for port 8001"
}
```

##### Invalid Request (Missing Fields)
**Input:**
```json
{
    "port": 8001
}
```

**Response:**
- Status Code: 400 Bad Request
- Body:
```json
{
    "error": "Invalid request payload"
}
```

##### Invalid Request (Virtual Service Not Found)
**Input:**
```json
{
    "port": 9999,
    "rate_limit": 15,
    "status_code": 429,
    "message": "Rate limit exceeded"
}
```

**Response:**
- Status Code: 404 Not Found
- Body:
```json
{
    "error": "Virtual Service not found"
}
```

#### 2. Rate Limited Request Behavior

##### When Rate Limit is Exceeded:
**Request:**
```http
GET /access/vs/8001
Authorization: Basic YmFsOjJmb3VyYWxs
```

**Response:**
- Status Code: 429 Too Many Requests
- Body:
```json
{
    "error": "Rate limit exceeded - Please try again later."
}
```

##### When Within Rate Limit:
**Request:**
```http
GET /access/vs/8001
Authorization: Basic YmFsOjJmb3VyYWxs
```

**Response:**
- Status Code: 200 OK
- Body: (Virtual service response)

#### 3. Authentication Requirements

##### Missing Authentication
**Request:**
```http
POST /access/vs/rate-limits
```

**Response:**
- Status Code: 401 Unauthorized
- Headers:
  ```
  WWW-Authenticate: Basic realm="Restricted"
  ```
- Body:
```json
{
    "error": "Unauthorized: Missing or invalid Authorization header"
}
```

##### Invalid Credentials
**Request:**
```http
POST /access/vs/rate-limits
Authorization: Basic aW52YWxpZDppbnZhbGlk
```

**Response:**
- Status Code: 401 Unauthorized
- Headers:
  ```
  WWW-Authenticate: Basic realm="Restricted"
  ```
- Body:
```json
{
    "error": "Unauthorized: Invalid credentials"
}
```

#### 4. Redis Error Handling

##### Redis Unavailable/Connection Timeout
**Request:**
```http
GET /access/vs/8001
Authorization: Basic YmFsOjJmb3VyYWxs
```

**Response:**
- Status Code: 200 OK
(System continues to operate with fallback behavior)

NOTE: go module name should be load-balancer for this project
