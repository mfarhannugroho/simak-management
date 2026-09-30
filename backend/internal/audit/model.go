package audit

import (
	"time"

	"github.com/google/uuid"
)

// Actions
const (
	ActionLogin               = "LOGIN"
	ActionLogout              = "LOGOUT"
	ActionSimakCreated        = "SIMAK_CREATED"
	ActionSimakSubmitted      = "SIMAK_SUBMITTED"
	ActionSimakApproved       = "SIMAK_APPROVED"
	ActionSimakRejected       = "SIMAK_REJECTED"
	ActionSimakRevision       = "SIMAK_REVISION_REQUESTED"
	ActionSimakIssued         = "SIMAK_ISSUED"
	ActionSimakCancelled      = "SIMAK_CANCELLED"
	ActionInspectionRecorded  = "INSPECTION_RECORDED"
	ActionDiscrepancyCreated  = "DISCREPANCY_CREATED"
	ActionDiscrepancyReview   = "DISCREPANCY_REVIEW_STARTED"
	ActionDiscrepancyResolved = "DISCREPANCY_RESOLVED"
)

// Entity Types
const (
	EntityAuth        = "AUTH"
	EntitySimak       = "SIMAK"
	EntityInspection  = "INSPECTION"
	EntityDiscrepancy = "DISCREPANCY"
	EntityUser        = "USER"
)

type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID     *uuid.UUID `gorm:"type:uuid" json:"user_id"`
	UserEmail  string     `gorm:"type:varchar(150)" json:"user_email"`
	UserName   string     `gorm:"type:varchar(150)" json:"user_name"`
	Action     string     `gorm:"type:varchar(60);not null" json:"action"`
	EntityType string     `gorm:"type:varchar(60);not null" json:"entity_type"`
	EntityID   string     `gorm:"type:varchar(100)" json:"entity_id"`
	Details    string     `gorm:"type:jsonb" json:"details"`
	IPAddress  string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  string     `gorm:"type:text" json:"user_agent"`
	CreatedAt  time.Time  `gorm:"default:now();not null" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type CreateAuditInput struct {
	UserID     *uuid.UUID
	UserEmail  string
	UserName   string
	Action     string
	EntityType string
	EntityID   string
	Details    interface{}
	IPAddress  string
	UserAgent  string
}

type AuditFilterParams struct {
	Action     string `form:"action"`
	EntityType string `form:"entity_type"`
	UserID     string `form:"user_id"`
	Search     string `form:"search"`
	DateFrom   string `form:"date_from"`
	DateTo     string `form:"date_to"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}
