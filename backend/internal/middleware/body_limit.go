package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySizeMiddleware limits the maximum size of incoming request bodies
// to prevent memory exhaustion and denial of service attacks.
func MaxBodySizeMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
