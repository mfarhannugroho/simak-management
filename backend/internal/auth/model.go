package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authentication account.
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string     `gorm:"type:varchar(150);unique;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	IsActive     bool       `gorm:"default:true;not null" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Roles  []Role  `gorm:"many2many:user_roles;" json:"roles"`
	Worker *Worker `gorm:"foreignKey:UserID" json:"worker"`
}

// Role represents a system role (Admin, Applicant, Approver, etc.).
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name        string       `gorm:"type:varchar(50);unique;not null" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission represents a granular capability (e.g. simak.approve).
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Worker represents the profile data for an individual worker or contractor.
type Worker struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID       uuid.UUID   `gorm:"type:uuid;unique;not null" json:"user_id"`
	CompanyID    *uuid.UUID  `gorm:"type:uuid" json:"company_id"`
	DepartmentID *uuid.UUID  `gorm:"type:uuid" json:"department_id"`
	WorkerID     string      `gorm:"type:varchar(50);unique" json:"worker_id"`
	FullName     string      `gorm:"type:varchar(150);not null" json:"full_name"`
	Position     string      `gorm:"type:varchar(100)" json:"position"`
	Phone        string      `gorm:"type:varchar(30)" json:"phone"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Company      *Company    `gorm:"foreignKey:CompanyID" json:"company"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department"`
}

// Company represents employer or partner vendor.
type Company struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

// Department represents INKA plant internal department.
type Department struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}
