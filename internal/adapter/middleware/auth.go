package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/gobank/internal/pkg/apperror"
	"github.com/yourusername/gobank/internal/pkg/token"
)

const (
	AuthorizationHeader = "Authorization"
	AuthorizationType   = "Bearer"
	UserIDKey           = "user_id"
	UserEmailKey        = "user_email"
	UserRoleKey         = "user_role"
)

func Auth(jwtManager token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrUnauthorized,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != AuthorizationType {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrInvalidToken,
			})
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			if err == token.ErrExpiredToken {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": apperror.ErrTokenExpired,
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrInvalidToken,
			})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(UserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrUnauthorized,
			})
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": apperror.ErrInternalServer,
			})
			return
		}

		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": apperror.ErrForbidden,
		})
	}
}
