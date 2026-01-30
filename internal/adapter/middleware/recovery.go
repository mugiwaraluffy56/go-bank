package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gobank/internal/infrastructure/logger"
	"github.com/yourusername/gobank/internal/pkg/apperror"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get(RequestIDKey)

				log.Error().
					Interface("panic", r).
					Str("request_id", requestID.(string)).
					Str("stack", string(debug.Stack())).
					Msg("Panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": apperror.ErrInternalServer,
				})
			}
		}()
		c.Next()
	}
}
