package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const authRateLimitPrefix = "ratelimit:auth:"

func AuthRateLimit(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := authRateLimitPrefix + c.ClientIP()
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			return
		}
		if count == 1 {
			if err := rdb.Expire(ctx, key, window).Err(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
				return
			}
		}
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("too many requests — limit is %d per minute", limit),
			})
			return
		}
		c.Next()
	}
}
