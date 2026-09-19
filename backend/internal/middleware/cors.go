package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures standard, robust CORS headers using gin-contrib/cors.
func CORSMiddleware(allowedOrigin string) gin.HandlerFunc {
	origins := make([]string, 0)
	for _, o := range strings.Split(allowedOrigin, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" && trimmed != "*" {
			origins = append(origins, trimmed)
		}
	}

	// Always ensure standard development and production origins are allowed
	defaultOrigins := []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:5174",
		"http://localhost:3000",
		"https://live-poll-ashy.vercel.app",
	}
	for _, defO := range defaultOrigins {
		found := false
		for _, o := range origins {
			if o == defO {
				found = true
				break
			}
		}
		if !found {
			origins = append(origins, defO)
		}
	}

	config := cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"Accept",
			"Origin",
			"Cache-Control",
			"X-Requested-With",
			"x-voter-id",
			"X-Voter-ID",
			"x-session-id",
			"X-Session-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Voter-ID",
			"x-voter-id",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}
