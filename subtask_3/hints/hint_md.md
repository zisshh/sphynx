Use Redis sorted sets with timestamps as scores to implement a sliding window. Here's a basic approach:

1. Use a key format like `rate_limit:{port}` for each virtual service
2. Store request timestamps in the sorted set
3. When checking the rate limit:
   - Remove old entries (older than 60 seconds)
   - Count remaining entries
   - Add new timestamp if within limit

Basic Redis commands you'll need:
```redis
ZREMRANGEBYSCORE key -inf min_timestamp
ZCARD key
ZADD key timestamp "request_id"
```

Remember to make these operations atomic using a Lua script.