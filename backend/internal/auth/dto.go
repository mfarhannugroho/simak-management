package auth

import "time"

// LoginRequest defines credentials for authentication.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest defines payload to obtain a new access token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest defines parameters for updating credentials.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// TokenPair holds generated access & refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// UserProfileDTO represents sanitized profile info returned to client.
type UserProfileDTO struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	FullName    string   `json:"full_name"`
	WorkerID    string   `json:"worker_id"`
	Position    string   `json:"position"`
	Phone       string   `json:"phone"`
	Company     string   `json:"company"`
	Department  string   `json:"department"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// LoginResponse contains token pair and initial profile data.
type LoginResponse struct {
	Tokens TokenPair      `json:"tokens"`
	User   UserProfileDTO `json:"user"`
}
