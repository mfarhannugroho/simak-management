package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"simak-backend/pkg/response"
)

// Handler exposes HTTP handlers for the auth module.
type Handler struct {
	service *Service
}

// NewHandler creates a new auth Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login handles user login with email and password.
// @Summary User Login
// @Description Authenticates credentials and returns JWT token pair
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data login tidak valid", err.Error())
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserInactive) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "Terjadi kesalahan sistem saat memproses login: "+err.Error())
		return
	}

	response.Success(c, "Login berhasil, selamat datang di SIMAK PT INKA", res)
}

// RefreshToken handles access token regeneration using refresh token.
// @Summary Refresh Token
// @Description Exchanges a valid refresh token for a new token pair
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} response.Response{data=TokenPair}
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Refresh token wajib disertakan", err.Error())
		return
	}

	tokens, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, "Token berhasil diperbarui", tokens)
}

// Logout revokes active JWT token in Redis blacklist.
// @Summary User Logout
// @Description Invalidates the active token session
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		response.Unauthorized(c, "Sesi tidak ditemukan")
		return
	}

	claims, ok := claimsVal.(*JWTClaims)
	if !ok {
		response.Unauthorized(c, "Format sesi tidak valid")
		return
	}

	_ = h.service.Logout(c.Request.Context(), claims)
	response.Success(c, "Logout berhasil, sesi telah diakhiri", nil)
}

// GetMe returns current logged-in user profile with roles and permissions.
// @Summary Current User Profile
// @Description Returns authenticated user identity, role, and permission matrix
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=UserProfileDTO}
// @Failure 401 {object} response.Response
// @Router /auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}

	profile, err := h.service.GetMe(c.Request.Context(), userID.(string))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Data profil pengguna berhasil dimuat", profile)
}

// ChangePassword allows authenticated user to update their password.
// @Summary Change Password
// @Description Updates password for active user
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Change Password Payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /auth/change-password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Pengguna tidak terautentikasi")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data kata sandi tidak valid", err.Error())
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID.(string), req); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Kata sandi berhasil diperbarui", nil)
}

// ListUsers handles user directory retrieval for Admin.
func (h *Handler) ListUsers(c *gin.Context) {
	search := c.Query("search")
	role := c.Query("role")
	users, err := h.service.ListUsers(c.Request.Context(), search, role)
	if err != nil {
		response.InternalServerError(c, "Gagal mengambil daftar pengguna: "+err.Error())
		return
	}
	response.Success(c, "Daftar pengguna berhasil dimuat", users)
}

// UpdateUserRole updates user's role assignment.
func (h *Handler) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID pengguna tidak valid", nil)
		return
	}
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Role wajib ditentukan", err.Error())
		return
	}
	if err := h.service.UpdateUserRole(c.Request.Context(), uid, req.Role); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Role pengguna berhasil diperbarui", nil)
}

// ToggleUserStatus activates or suspends a user account.
func (h *Handler) ToggleUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID pengguna tidak valid", nil)
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Status tidak valid", err.Error())
		return
	}
	if err := h.service.ToggleUserStatus(c.Request.Context(), uid, req.IsActive); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	msg := "Akun pengguna berhasil diaktifkan"
	if !req.IsActive {
		msg = "Akun pengguna berhasil dinonaktifkan"
	}
	response.Success(c, msg, nil)
}

// ListRoles returns all roles and their permissions.
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Gagal mengambil daftar role: "+err.Error())
		return
	}
	response.Success(c, "Daftar role dan hak akses berhasil dimuat", roles)
}
