package inspection

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/auth"
	"simak-backend/internal/simak"
)

// Inspection represents a physical security check event at the PT INKA gate
type Inspection struct {
	ID                   uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SimakID              uuid.UUID        `gorm:"type:uuid;not null" json:"simak_id"`
	InspectorID          uuid.UUID        `gorm:"type:uuid;not null" json:"inspector_id"`
	GateName             string           `gorm:"type:varchar(100);default:'Pos Gerbang Utama PT INKA';not null" json:"gate_name"`
	InspectionType       string           `gorm:"type:varchar(30);default:'MASUK';not null" json:"inspection_type"` // "MASUK", "KELUAR"
	Status               string           `gorm:"type:varchar(30);default:'MATCH';not null" json:"status"`           // "MATCH", "MISMATCH", "INCOMPLETE"
	TotalRegisteredItems int              `gorm:"default:0;not null" json:"total_registered_items"`
	TotalActualItems     int              `gorm:"default:0;not null" json:"total_actual_items"`
	Notes                string           `gorm:"type:text" json:"notes"`
	InspectedAt          time.Time        `gorm:"default:now();not null" json:"inspected_at"`
	CreatedAt            time.Time        `json:"created_at"`

	// Relations
	Simak          *simak.Simak     `gorm:"foreignKey:SimakID" json:"simak,omitempty"`
	Inspector      *auth.User       `gorm:"foreignKey:InspectorID" json:"inspector,omitempty"`
	Items          []InspectionItem `gorm:"foreignKey:InspectionID" json:"items,omitempty"`
}

func (Inspection) TableName() string {
	return "inspections"
}

// InspectionItem represents an individual line item verified during physical inspection
type InspectionItem struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	InspectionID uuid.UUID  `gorm:"type:uuid;not null" json:"inspection_id"`
	SimakItemID  *uuid.UUID `gorm:"type:uuid" json:"simak_item_id"`
	ItemName     string     `gorm:"type:varchar(150);not null" json:"item_name"`
	Brand        string     `gorm:"type:varchar(100)" json:"brand"`
	SerialNumber string     `gorm:"type:varchar(100)" json:"serial_number"`
	ExpectedQty  int        `gorm:"default:1;not null" json:"expected_qty"`
	ActualQty    int        `gorm:"default:0;not null" json:"actual_qty"`
	Unit         string     `gorm:"type:varchar(30);default:'Unit';not null" json:"unit"`
	Status       string     `gorm:"type:varchar(30);default:'MATCH';not null" json:"status"` // "MATCH", "MISMATCH", "UNREGISTERED"
	Notes        string     `gorm:"type:text" json:"notes"`
	PhotoURL     string     `gorm:"type:text" json:"photo_url"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (InspectionItem) TableName() string {
	return "inspection_items"
}
