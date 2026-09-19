package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware provides centralized panic recovery and returns a standardized 500 JSON response.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				errStr := fmt.Sprintf("%v", r)
				log.Printf("[ERROR] Panic recovered: %s", errStr)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "An unexpected internal server error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}
