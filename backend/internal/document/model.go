package document

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/auth"
	"simak-backend/internal/simak"
)

// SimakDocument represents an official issued document for a SIMAK application
type SimakDocument struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SimakID        uuid.UUID    `gorm:"type:uuid;not null" json:"simak_id"`
	DocumentNumber string       `gorm:"type:varchar(100);not null" json:"document_number"`
	FileName       string       `gorm:"type:varchar(255);not null" json:"file_name"`
	FileURL        string       `gorm:"type:text" json:"file_url"`
	DocType        string       `gorm:"type:varchar(50);default:'OFFICIAL_PDF';not null" json:"doc_type"`
	QRCodeData     string       `gorm:"type:text" json:"qr_code_data"`
	IssuedBy       *uuid.UUID   `gorm:"type:uuid" json:"issued_by"`
	IssuedAt       time.Time    `json:"issued_at"`
	CreatedAt      time.Time    `json:"created_at"`

	// Relations
	Simak          *simak.Simak `gorm:"foreignKey:SimakID" json:"simak,omitempty"`
	Issuer         *auth.User   `gorm:"foreignKey:IssuedBy" json:"issuer,omitempty"`
}

func (SimakDocument) TableName() string {
	return "simak_documents"
}
