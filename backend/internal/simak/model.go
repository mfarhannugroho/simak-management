package simak

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/item"
	"simak-backend/internal/worker"
)

// SimakStatus constants defining the lifecycle of a SIMAK request
const (
	StatusDraft            = "DRAFT"
	StatusSubmitted        = "SUBMITTED"
	StatusUnderReview      = "UNDER_REVIEW"
	StatusRequiresRevision = "REQUIRES_REVISION"
	StatusRejected         = "REJECTED"
	StatusApproved         = "APPROVED"
	StatusIssued           = "ISSUED"
	StatusActive           = "ACTIVE"
	StatusVerified         = "VERIFIED"
	StatusCompleted        = "COMPLETED"
	StatusCancelled        = "CANCELLED"
	StatusExpired          = "EXPIRED"
)

// Direction constants
const (
	DirectionMasuk       = "MASUK"
	DirectionKeluar      = "KELUAR"
	DirectionMasukKeluar = "MASUK_KELUAR"
)

// Simak represents a worker carried equipment permit
type Simak struct {
	ID              uuid.UUID         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ApplicantID     uuid.UUID         `gorm:"column:worker_id;not null" json:"worker_id"`
	SimakNumber     *string           `gorm:"type:varchar(60);unique" json:"simak_number"`
	Purpose         string            `gorm:"type:text;not null" json:"purpose"`
	Direction       string            `gorm:"type:varchar(30);default:'MASUK';not null" json:"direction"`
	DestinationArea string            `gorm:"type:varchar(150);not null" json:"destination_area"`
	StartDate       time.Time         `gorm:"type:date;not null" json:"start_date"`
	EndDate         time.Time         `gorm:"type:date;not null" json:"end_date"`
	Status          string            `gorm:"type:varchar(40);default:'DRAFT';not null" json:"status"`
	Notes           string            `gorm:"type:text" json:"notes"`
	RevisionNotes   string            `gorm:"type:text" json:"revision_notes"`
	RejectionReason string            `gorm:"type:text" json:"rejection_reason"`
	QRToken         *string           `gorm:"type:varchar(100);unique" json:"qr_token"`
	IssuedAt        *time.Time        `json:"issued_at"`
	IssuedBy        *uuid.UUID        `gorm:"type:uuid" json:"issued_by"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`

	// Relations
	WorkerProfile *worker.Worker    `gorm:"foreignKey:ApplicantID;references:ID" json:"worker,omitempty"`
	Items         []SimakItem       `gorm:"foreignKey:SimakID" json:"items,omitempty"`
	Attachments   []SimakAttachment `gorm:"foreignKey:SimakID" json:"attachments,omitempty"`
}

// SimakItem represents a piece of equipment/tool included in a SIMAK permit
type SimakItem struct {
	ID           uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SimakID      uuid.UUID          `gorm:"type:uuid;not null" json:"simak_id"`
	CategoryID   *uuid.UUID         `gorm:"type:uuid" json:"category_id"`
	ItemName     string             `gorm:"type:varchar(150);not null" json:"item_name"`
	Brand        string             `gorm:"type:varchar(100)" json:"brand"`
	Model        string             `gorm:"type:varchar(100)" json:"model"`
	SerialNumber string             `gorm:"type:varchar(100)" json:"serial_number"`
	SizeSpec     string             `gorm:"type:varchar(100)" json:"size_spec"`
	Quantity     int                `gorm:"default:1;not null" json:"quantity"`
	Unit         string             `gorm:"type:varchar(30);default:'Unit';not null" json:"unit"`
	Description  string             `gorm:"type:text" json:"description"`
	PhotoURL     string             `gorm:"type:text" json:"photo_url"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`

	// Relations
	Category *item.ItemCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// SimakAttachment represents an attached official assignment letter or credential document
type SimakAttachment struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SimakID        uuid.UUID `gorm:"type:uuid;not null" json:"simak_id"`
	FileName       string    `gorm:"type:varchar(255);not null" json:"file_name"`
	FileURL        string    `gorm:"type:text;not null" json:"file_url"`
	FileType       string    `gorm:"type:varchar(100)" json:"file_type"`
	FileSize       int64     `json:"file_size"`
	AttachmentType string    `gorm:"type:varchar(50);default:'SURAT_TUGAS';not null" json:"attachment_type"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Simak) TableName() string {
	return "simak"
}

func (SimakItem) TableName() string {
	return "simak_items"
}

func (SimakAttachment) TableName() string {
	return "simak_attachments"
}
