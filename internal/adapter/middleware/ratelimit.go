package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gobank/internal/adapter/repository/redis"
	"github.com/yourusername/gobank/internal/pkg/apperror"
)

func RateLimit(limiter *redis.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if userID, exists := c.Get(UserIDKey); exists {
			key = fmt.Sprintf("user:%v", userID)
		}

		allowed, remaining, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.GetLimit()))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": apperror.ErrTooManyRequests,
			})
			return
		}

		c.Next()
	}
}

func RateLimitByIP(limiter *redis.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ip:%s", c.ClientIP())

		allowed, remaining, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.GetLimit()))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": apperror.ErrTooManyRequests,
			})
			return
		}

		c.Next()
	}
}
