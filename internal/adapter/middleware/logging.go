package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/infrastructure/logger"
)

const RequestIDKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func Logging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		requestID, _ := c.Get(RequestIDKey)

		logEvent := log.Info()
		if statusCode >= 500 {
			logEvent = log.Error()
		} else if statusCode >= 400 {
			logEvent = log.Warn()
		}

		logEvent.
			Str("request_id", requestID.(string)).
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Msg("HTTP request")
	}
}
