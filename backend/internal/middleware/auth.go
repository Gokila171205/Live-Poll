package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"live-poll-backend/internal/models"
	"live-poll-backend/internal/services"
)

// AuthMiddleware creates a Gin middleware that validates Bearer JWT tokens.
func AuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			abortWithAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format. Expected 'Bearer <token>'")
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := authService.ValidateToken(tokenStr)
		if err != nil {
			abortWithAuthError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			return
		}

		// Store authenticated user ID extracted securely from token
		c.Set(models.ContextKeyUserID, claims.UserID)
		c.Next()
	}
}

// GetAuthenticatedUserID extracts the verified user ID from the Gin request context.
func GetAuthenticatedUserID(c *gin.Context) (string, bool) {
	val, exists := c.Get(models.ContextKeyUserID)
	if !exists {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok && userID != ""
}

// OptionalAuthMiddleware inspects the Authorization header if present and stores the verified user ID without aborting unauthenticated requests.
func OptionalAuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr := strings.TrimSpace(parts[1])
				if tokenStr != "" && authService != nil {
					if claims, err := authService.ValidateToken(tokenStr); err == nil && claims != nil && claims.UserID != "" {
						c.Set(models.ContextKeyUserID, claims.UserID)
					}
				}
			}
		}
		c.Next()
	}
}

func abortWithAuthError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
