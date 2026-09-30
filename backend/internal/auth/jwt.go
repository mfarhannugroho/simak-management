package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"simak-backend/config"
)

var (
	ErrInvalidToken = errors.New("token tidak valid atau sudah kedaluwarsa")
	ErrBlacklisted  = errors.New("token telah dicabut (logged out)")
)

// JWTClaims holds authenticated user identity in the JWT payload.
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   string   `json:"token_type"` // "access" | "refresh"
	jwt.RegisteredClaims
}

// GenerateTokenPair creates both Access Token and Refresh Token for the user.
func GenerateTokenPair(user *User, roles []string, permissions []string, cfg *config.Config) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(time.Duration(cfg.JWT.AccessExpiry) * time.Hour)
	refreshExpiry := now.Add(time.Duration(cfg.JWT.RefreshExpiry) * 24 * time.Hour)

	accessJTI := uuid.New().String()
	accessClaims := JWTClaims{
		UserID:      user.ID.String(),
		Email:       user.Email,
		Roles:       roles,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			Subject:   user.ID.String(),
			Issuer:    cfg.App.Name,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString([]byte(cfg.JWT.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani access token: %w", err)
	}

	refreshJTI := uuid.New().String()
	refreshClaims := JWTClaims{
		UserID:    user.ID.String(),
		Email:     user.Email,
		Roles:     roles,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			Subject:   user.ID.String(),
			Issuer:    cfg.App.Name,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(cfg.JWT.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// ParseToken parses and validates a JWT token string against the specified secret.
func ParseToken(tokenStr string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// BlacklistToken stores a token JTI in Redis until its natural expiration.
func BlacklistToken(ctx context.Context, rdb *redis.Client, jti string, remaining time.Duration) error {
	if rdb == nil || jti == "" {
		return nil
	}
	if remaining <= 0 {
		return nil
	}
	key := fmt.Sprintf("blacklist:token:%s", jti)
	return rdb.Set(ctx, key, "1", remaining).Err()
}

// IsTokenBlacklisted checks if the given JTI has been revoked in Redis.
func IsTokenBlacklisted(ctx context.Context, rdb *redis.Client, jti string) bool {
	if rdb == nil || jti == "" {
		return false
	}
	key := fmt.Sprintf("blacklist:token:%s", jti)
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return val == "1"
}
