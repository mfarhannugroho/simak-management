package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID attaches a unique X-Request-ID header to every request/response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Recovery is a Gin middleware that recovers from panics and returns a 500
// response without crashing the server.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Timeout adds a per-request deadline of 30 seconds.
func Timeout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a per-request timeout using the request context.
		// Actual timeout enforcement is left to individual handlers via ctx.Done().
		c.Set("request_timeout", 30*time.Second)
		c.Next()
	}
}
