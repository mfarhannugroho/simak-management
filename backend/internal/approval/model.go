package approval

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/company"
	"simak-backend/internal/department"
)

// ApprovalMatrix represents a configurable approval route rule
type ApprovalMatrix struct {
	ID           uuid.UUID              `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name         string                 `gorm:"type:varchar(150);not null" json:"name"`
	Description  string                 `gorm:"type:text" json:"description"`
	DepartmentID *uuid.UUID             `gorm:"type:uuid" json:"department_id"`
	CompanyID    *uuid.UUID             `gorm:"type:uuid" json:"company_id"`
	IsActive     bool                   `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`

	// Relations
	Department *department.Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Company    *company.Company       `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Steps      []ApprovalMatrixStep   `gorm:"foreignKey:MatrixID;order:step_order asc" json:"steps,omitempty"`
}

func (ApprovalMatrix) TableName() string {
	return "approval_matrices"
}

// ApprovalMatrixStep represents an individual step in the approval route
type ApprovalMatrixStep struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MatrixID       uuid.UUID `gorm:"type:uuid;not null" json:"matrix_id"`
	StepOrder      int       `gorm:"not null" json:"step_order"`
	StepName       string    `gorm:"type:varchar(100);not null" json:"step_name"`
	RequiredRole   string    `gorm:"type:varchar(50);default:'Approver';not null" json:"required_role"`
	SpecificUserID *uuid.UUID `gorm:"type:uuid" json:"specific_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}

func (ApprovalMatrixStep) TableName() string {
	return "approval_matrix_steps"
}

// SimakApproval represents an actual instantiated approval stage for a specific SIMAK application
type SimakApproval struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SimakID         uuid.UUID  `gorm:"type:uuid;not null" json:"simak_id"`
	StepOrder       int        `gorm:"not null" json:"step_order"`
	StepName        string     `gorm:"type:varchar(100);not null" json:"step_name"`
	RequiredRole    string     `gorm:"type:varchar(50);default:'Approver';not null" json:"required_role"`
	AssignedUserID  *uuid.UUID `gorm:"type:uuid" json:"assigned_user_id"`
	Status          string     `gorm:"type:varchar(30);default:'PENDING';not null" json:"status"`
	ActionByUserID  *uuid.UUID `gorm:"type:uuid" json:"action_by_user_id"`
	ActionAt        *time.Time `json:"action_at"`
	Comments        string     `gorm:"type:text" json:"comments"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relational Actor Summary
	ActionByName string `gorm:"-" json:"action_by_name,omitempty"`
}

func (SimakApproval) TableName() string {
	return "simak_approvals"
}
