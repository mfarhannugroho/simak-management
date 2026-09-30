package discrepancy

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/auth"
	"simak-backend/internal/inspection"
	"simak-backend/internal/simak"
)

const (
	TypeQuantityMismatch        = "QUANTITY_MISMATCH"
	TypeUnregisteredItem        = "UNREGISTERED_ITEM"
	TypeMissingItem             = "MISSING_ITEM"
	TypeItemDescriptionMismatch = "ITEM_DESCRIPTION_MISMATCH"
	TypeOther                   = "OTHER"

	StatusOpen     = "OPEN"
	StatusInReview = "IN_REVIEW"
	StatusResolved = "RESOLVED"
	StatusClosed   = "CLOSED"

	ActionAllowEntry       = "ALLOW_ENTRY"
	ActionDenyEntry        = "DENY_ENTRY"
	ActionConfiscateHold   = "CONFISCATE_HOLD_AT_GATE"
	ActionReturnToVendor   = "RETURN_TO_VENDOR"
	ActionRevisePermit     = "REVISE_PERMIT"
	ActionWarningIssued    = "WARNING_ISSUED"
)

type Discrepancy struct {
	ID                uuid.UUID              `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	DiscrepancyNumber string                 `gorm:"type:varchar(50);not null;unique" json:"discrepancy_number"`
	SimakID           uuid.UUID              `gorm:"type:uuid;not null;index" json:"simak_id"`
	InspectionID      uuid.UUID              `gorm:"type:uuid;not null;index" json:"inspection_id"`
	InspectionItemID  *uuid.UUID             `gorm:"type:uuid" json:"inspection_item_id"`
	ItemName          string                 `gorm:"type:varchar(255);not null" json:"item_name"`
	Brand             string                 `gorm:"type:varchar(100)" json:"brand"`
	SerialNumber      string                 `gorm:"type:varchar(100)" json:"serial_number"`
	DiscrepancyType   string                 `gorm:"type:varchar(50);not null;index" json:"discrepancy_type"`
	ExpectedQty       int                    `gorm:"type:int;not null;default:0" json:"expected_qty"`
	ActualQty         int                    `gorm:"type:int;not null;default:0" json:"actual_qty"`
	Unit              string                 `gorm:"type:varchar(50);not null;default:'Unit'" json:"unit"`
	Description       string                 `gorm:"type:text" json:"description"`
	EvidencePhotoURL  string                 `gorm:"type:text" json:"evidence_photo_url"`
	ReportedBy        uuid.UUID              `gorm:"type:uuid;not null" json:"reported_by"`
	ReviewedBy        *uuid.UUID             `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt        *time.Time             `gorm:"type:timestamptz" json:"reviewed_at"`
	ResolutionAction  string                 `gorm:"type:varchar(50)" json:"resolution_action"`
	ResolutionNotes   string                 `gorm:"type:text" json:"resolution_notes"`
	Status            string                 `gorm:"type:varchar(50);not null;default:'OPEN';index" json:"status"`
	CreatedAt         time.Time              `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt         time.Time              `gorm:"type:timestamptz;not null;default:now()" json:"updated_at"`

	// Relations
	Simak          *simak.Simak           `gorm:"foreignKey:SimakID" json:"simak,omitempty"`
	Inspection     *inspection.Inspection `gorm:"foreignKey:InspectionID" json:"inspection,omitempty"`
	InspectionItem *inspection.InspectionItem `gorm:"foreignKey:InspectionItemID" json:"inspection_item,omitempty"`
	Reporter       *auth.User             `gorm:"foreignKey:ReportedBy" json:"reporter,omitempty"`
	Reviewer       *auth.User             `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (Discrepancy) TableName() string {
	return "discrepancies"
}
