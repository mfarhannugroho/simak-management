package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"simak-backend/config"
)

var (
	ErrUserNotFound       = errors.New("pengguna dengan email tersebut tidak ditemukan")
	ErrInvalidCredentials = errors.New("email atau kata sandi tidak cocok")
	ErrUserInactive       = errors.New("akun pengguna telah dinonaktifkan, silakan hubungi administrator")
)

// Service defines authentication & user credential business logic.
type Service struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

// NewService constructs an instance of AuthService.
func NewService(db *gorm.DB, rdb *redis.Client, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		rdb: rdb,
		cfg: cfg,
	}
}

// Login authenticates a user by email & password.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var user User
	err := s.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("Worker.Company").
		Preload("Worker.Department").
		Where("LOWER(email) = LOWER(?)", req.Email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify Bcrypt password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Extract unique role names and permission names
	roles, permissions := s.extractRolesAndPermissions(&user)

	// Update last login timestamp
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.db.WithContext(ctx).Model(&user).Update("last_login_at", user.LastLoginAt)

	// Generate JWT Access & Refresh tokens
	tokens, err := GenerateTokenPair(&user, roles, permissions, s.cfg)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Tokens: *tokens,
		User:   s.toUserProfileDTO(&user, roles, permissions),
	}, nil
}

// RefreshToken produces a new access token using a valid refresh token.
func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenPair, error) {
	claims, err := ParseToken(req.RefreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	// Check if this refresh token was already revoked/used
	if IsTokenBlacklisted(ctx, s.rdb, claims.ID) {
		return nil, ErrBlacklisted
	}

	// Blacklist the old refresh token (rotation)
	if claims.ExpiresAt != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		_ = BlacklistToken(ctx, s.rdb, claims.ID, remaining)
	}

	// Fetch user to confirm still active and get fresh permissions
	var user User
	err = s.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Where("id = ?", claims.UserID).
		First(&user).Error
	if err != nil || !user.IsActive {
		return nil, ErrInvalidToken
	}

	roles, permissions := s.extractRolesAndPermissions(&user)
	return GenerateTokenPair(&user, roles, permissions, s.cfg)
}

// Logout invalidates the active access token in Redis blacklist.
func (s *Service) Logout(ctx context.Context, claims *JWTClaims) error {
	if claims == nil || claims.ID == "" {
		return nil
	}
	if claims.ExpiresAt != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		return BlacklistToken(ctx, s.rdb, claims.ID, remaining)
	}
	return nil
}

// GetMe retrieves the current authenticated user's detailed profile.
func (s *Service) GetMe(ctx context.Context, userID string) (*UserProfileDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var user User
	err = s.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("Worker.Company").
		Preload("Worker.Department").
		Where("id = ?", uid).
		First(&user).Error
	if err != nil {
		return nil, ErrUserNotFound
	}

	roles, permissions := s.extractRolesAndPermissions(&user)
	dto := s.toUserProfileDTO(&user, roles, permissions)
	return &dto, nil
}

// ChangePassword updates the user's password after verifying the current password.
func (s *Service) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidToken
	}

	var user User
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&user).Error; err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("kata sandi lama tidak cocok")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.db.WithContext(ctx).Model(&user).Update("password_hash", string(newHash)).Error
}

func (s *Service) extractRolesAndPermissions(user *User) ([]string, []string) {
	rolesMap := make(map[string]bool)
	permMap := make(map[string]bool)

	for _, role := range user.Roles {
		rolesMap[role.Name] = true
		for _, perm := range role.Permissions {
			permMap[perm.Name] = true
		}
	}

	roles := make([]string, 0, len(rolesMap))
	for r := range rolesMap {
		roles = append(roles, r)
	}

	perms := make([]string, 0, len(permMap))
	for p := range permMap {
		perms = append(perms, p)
	}

	return roles, perms
}

func (s *Service) toUserProfileDTO(user *User, roles []string, perms []string) UserProfileDTO {
	dto := UserProfileDTO{
		ID:          user.ID.String(),
		Email:       user.Email,
		Roles:       roles,
		Permissions: perms,
	}

	if user.Worker != nil {
		dto.FullName = user.Worker.FullName
		dto.WorkerID = user.Worker.WorkerID
		dto.Position = user.Worker.Position
		dto.Phone = user.Worker.Phone
		if user.Worker.Company != nil {
			dto.Company = user.Worker.Company.Name
		}
		if user.Worker.Department != nil {
			dto.Department = user.Worker.Department.Name
		}
	}

	return dto
}

// AdminUserItem represents a user in the admin management table.
type AdminUserItem struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Roles       []string   `json:"roles"`
	WorkerID    string     `json:"worker_id,omitempty"`
	FullName    string     `json:"full_name,omitempty"`
	Position    string     `json:"position,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	CompanyName string     `json:"company_name,omitempty"`
	Department  string     `json:"department_name,omitempty"`
}

// ListUsers retrieves all registered user accounts with worker profile & role information.
func (s *Service) ListUsers(ctx context.Context, search, roleFilter string) ([]AdminUserItem, error) {
	var users []User
	query := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Worker.Company").
		Preload("Worker.Department").
		Order("created_at ASC")

	if search != "" {
		like := "%" + search + "%"
		query = query.Joins("LEFT JOIN workers ON workers.user_id = users.id").
			Where("users.email ILIKE ? OR workers.full_name ILIKE ? OR workers.worker_id ILIKE ?", like, like, like)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	result := make([]AdminUserItem, 0, len(users))
	for _, u := range users {
		var roleNames []string
		hasMatchingRole := roleFilter == ""
		for _, r := range u.Roles {
			roleNames = append(roleNames, r.Name)
			if roleFilter != "" && r.Name == roleFilter {
				hasMatchingRole = true
			}
		}
		if !hasMatchingRole {
			continue
		}

		item := AdminUserItem{
			ID:          u.ID,
			Email:       u.Email,
			IsActive:    u.IsActive,
			LastLoginAt: u.LastLoginAt,
			CreatedAt:   u.CreatedAt,
			Roles:       roleNames,
		}
		if u.Worker != nil {
			item.WorkerID = u.Worker.WorkerID
			item.FullName = u.Worker.FullName
			item.Position = u.Worker.Position
			item.Phone = u.Worker.Phone
			if u.Worker.Company != nil {
				item.CompanyName = u.Worker.Company.Name
			}
			if u.Worker.Department != nil {
				item.Department = u.Worker.Department.Name
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// UpdateUserRole reassigns a user to a different system role.
func (s *Service) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	var role Role
	if err := s.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("role '%s' tidak ditemukan", roleName)
	}
	var user User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return errors.New("pengguna tidak ditemukan")
	}
	return s.db.WithContext(ctx).Model(&user).Association("Roles").Replace(&role)
}

// ToggleUserStatus activates or suspends a user account.
func (s *Service) ToggleUserStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	return s.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Update("is_active", isActive).Error
}

// ListRoles returns all roles and their permissions.
func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	err := s.db.WithContext(ctx).Preload("Permissions").Order("name ASC").Find(&roles).Error
	return roles, err
}
