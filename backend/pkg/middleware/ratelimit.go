package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"simak-backend/pkg/response"
)

// RateLimiter creates a Redis-backed rate limiting middleware
// limit: maximum requests allowed within windowDuration
func RateLimiter(rdb *redis.Client, limit int64, windowDuration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Identify client by IP or authenticated user_id
		identifier := c.ClientIP()
		if uid, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", uid)
		}

		key := fmt.Sprintf("ratelimit:%s:%s", c.FullPath(), identifier)
		ctx := context.Background()

		// Increment request counter in Redis
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// If Redis is temporarily unreachable, allow request to proceed (graceful degradation)
			c.Next()
			return
		}

		// On first request, set expiry for the window
		if count == 1 {
			rdb.Expire(ctx, key, windowDuration)
		}

		// Inject rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		// Check if threshold exceeded
		if count > limit {
			ttl, _ := rdb.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Response{
				Success: false,
				Message: fmt.Sprintf("Terlalu banyak permintaan. Silakan tunggu %d detik sebelum mencoba kembali.", int(ttl.Seconds())),
			})
			return
		}

		c.Next()
	}
}
