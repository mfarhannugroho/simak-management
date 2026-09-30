package notification

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/auth"
)

const (
	TypeSimakSubmitted      = "SIMAK_SUBMITTED"
	TypeApprovalRequired    = "APPROVAL_REQUIRED"
	TypeSimakApproved       = "SIMAK_APPROVED"
	TypeSimakRejected       = "SIMAK_REJECTED"
	TypeRevisionRequired    = "REVISION_REQUIRED"
	TypeSimakIssued         = "SIMAK_ISSUED"
	TypeSimakExpiring       = "SIMAK_EXPIRING"
	TypeSimakExpired        = "SIMAK_EXPIRED"
	TypeInspectionCompleted = "INSPECTION_COMPLETED"
	TypeDiscrepancyCreated  = "DISCREPANCY_CREATED"
	TypeDiscrepancyResolved = "DISCREPANCY_RESOLVED"
)

type Notification struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Title         string     `gorm:"type:varchar(255);not null" json:"title"`
	Message       string     `gorm:"type:text;not null" json:"message"`
	Type          string     `gorm:"type:varchar(50);not null;index" json:"type"`
	ReferenceID   string     `gorm:"type:varchar(100)" json:"reference_id"`
	ReferenceType string     `gorm:"type:varchar(50)" json:"reference_type"` // simak, discrepancy, inspection
	ActionURL     string     `gorm:"type:varchar(255)" json:"action_url"`
	IsRead        bool       `gorm:"type:boolean;not null;default:false;index" json:"is_read"`
	ReadAt        *time.Time `gorm:"type:timestamptz" json:"read_at"`
	CreatedAt     time.Time  `gorm:"type:timestamptz;not null;default:now();index" json:"created_at"`

	// Relations
	User *auth.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}
