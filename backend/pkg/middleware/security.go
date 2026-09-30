package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersMiddleware adds OWASP recommended security headers to all responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME-sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent Clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Cross-Site Scripting protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy for secure API and SPA serving
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https: blob:; connect-src 'self' ws: wss: http: https:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline';")

		c.Next()
	}
}
