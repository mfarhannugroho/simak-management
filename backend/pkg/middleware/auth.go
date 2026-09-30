package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"simak-backend/config"
	"simak-backend/internal/auth"
	"simak-backend/pkg/response"
)

// AuthMiddleware validates JWT Bearer tokens and checks Redis blacklists.
func AuthMiddleware(cfg *config.Config, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Trim(strings.TrimSpace(c.GetHeader("Authorization")), "\"")
		if authHeader == "" {
			response.Unauthorized(c, "Header otorisasi tidak ditemukan")
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) < 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Format header otorisasi harus Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := auth.ParseToken(tokenStr, cfg.JWT.AccessSecret)
		if err != nil {
			response.Unauthorized(c, "Token otorisasi tidak valid atau telah kedaluwarsa")
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			response.Unauthorized(c, "Tipe token otorisasi tidak valid")
			c.Abort()
			return
		}

		// Check if token was revoked via Redis blacklist
		if auth.IsTokenBlacklisted(c.Request.Context(), rdb, claims.ID) {
			response.Unauthorized(c, "Sesi token ini telah diakhiri (silakan login kembali)")
			c.Abort()
			return
		}

		// Inject authenticated user context into request
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequirePermission enforces granular capability checks on routes.
func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permsVal, exists := c.Get("permissions")
		if !exists {
			response.Forbidden(c, "Akses ditolak: izin pengguna tidak ditemukan")
			c.Abort()
			return
		}

		permissions, ok := permsVal.([]string)
		if !ok {
			response.Forbidden(c, "Akses ditolak: format izin pengguna tidak valid")
			c.Abort()
			return
		}

		// Admin bypass: if user has 'Admin' role, permit all actions
		if rolesVal, rExists := c.Get("roles"); rExists {
			if roles, rOk := rolesVal.([]string); rOk {
				for _, r := range roles {
					if r == "Admin" {
						c.Next()
						return
					}
				}
			}
		}

		// Check granular permission list
		hasPermission := false
		for _, p := range permissions {
			if p == requiredPermission || p == "*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			response.Forbidden(c, "Akses ditolak: Anda tidak memiliki izin untuk aksi ini ("+requiredPermission+")")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole enforces specific role membership.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal, exists := c.Get("roles")
		if !exists {
			response.Forbidden(c, "Akses ditolak: peran pengguna tidak ditemukan")
			c.Abort()
			return
		}

		userRoles, ok := rolesVal.([]string)
		if !ok {
			response.Forbidden(c, "Akses ditolak: format peran pengguna tidak valid")
			c.Abort()
			return
		}

		hasRole := false
		for _, ur := range userRoles {
			if ur == "Admin" { // Admin can access all role-restricted routes
				hasRole = true
				break
			}
			for _, ar := range allowedRoles {
				if ur == ar {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			response.Forbidden(c, "Akses ditolak: Anda tidak memiliki peran yang diizinkan untuk mengakses halaman ini")
			c.Abort()
			return
		}

		c.Next()
	}
}
