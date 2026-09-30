package document

import (
	"time"

	"github.com/google/uuid"
)

type IssueSimakRequest struct {
	CustomNumber string `json:"custom_number"`
	Notes        string `json:"notes"`
}

type VerificationItemInfo struct {
	ItemName     string `json:"item_name"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	Quantity     int    `json:"quantity"`
	Unit         string `json:"unit"`
	Description  string `json:"description"`
}

type PublicVerificationResponse struct {
	Valid           bool                   `json:"valid"`
	Status          string                 `json:"status"` // VALID, EXPIRED, CANCELLED, NOT_YET_VALID, INVALID
	Reason          string                 `json:"reason,omitempty"`
	SimakNumber     string                 `json:"simak_number,omitempty"`
	ApplicantName   string                 `json:"applicant_name,omitempty"`
	WorkerID        string                 `json:"worker_id,omitempty"`
	Company         string                 `json:"company,omitempty"`
	Department      string                 `json:"department,omitempty"`
	Position        string                 `json:"position,omitempty"`
	Purpose         string                 `json:"purpose,omitempty"`
	Direction       string                 `json:"direction,omitempty"`
	DestinationArea string                 `json:"destination_area,omitempty"`
	StartDate       string                 `json:"start_date,omitempty"`
	EndDate         string                 `json:"end_date,omitempty"`
	IssuedAt        string                 `json:"issued_at,omitempty"`
	TotalItems      int                    `json:"total_items,omitempty"`
	Items           []VerificationItemInfo `json:"items,omitempty"`
}

type IssuanceQueueItem struct {
	ID              uuid.UUID  `json:"id"`
	SimakNumber     string     `json:"simak_number"`
	Purpose         string     `json:"purpose"`
	Direction       string     `json:"direction"`
	DestinationArea string     `json:"destination_area"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	Status          string     `json:"status"`
	QRToken         *string    `json:"qr_token,omitempty"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	WorkerName      string     `json:"worker_name"`
	WorkerID        string     `json:"worker_id"`
	Company         string     `json:"company"`
	Department      string     `json:"department"`
	TotalItems      int        `json:"total_items"`
}
