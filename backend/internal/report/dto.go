package report

import (
	"time"

	"github.com/google/uuid"
)

type SimakReportRow struct {
	ID              uuid.UUID  `json:"id"`
	SimakNumber     string     `json:"simak_number"`
	WorkerName      string     `json:"worker_name"`
	WorkerNIP       string     `json:"worker_nip"`
	DepartmentName  string     `json:"department_name"`
	CompanyName     string     `json:"company_name"`
	Direction       string     `json:"direction"`
	DestinationArea string     `json:"destination_area"`
	StartDate       string     `json:"start_date"`
	EndDate         string     `json:"end_date"`
	Status          string     `json:"status"`
	TotalItems      int        `json:"total_items"`
	IssuedAt        *time.Time `json:"issued_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type VerificationReportRow struct {
	ID                   uuid.UUID `json:"id"`
	SimakNumber          string    `json:"simak_number"`
	WorkerName           string    `json:"worker_name"`
	InspectorName        string    `json:"inspector_name"`
	GateName             string    `json:"gate_name"`
	InspectionType       string    `json:"inspection_type"`
	Status               string    `json:"status"`
	TotalRegisteredItems int       `json:"total_registered_items"`
	TotalActualItems     int       `json:"total_actual_items"`
	Notes                string    `json:"notes"`
	InspectedAt          time.Time `json:"inspected_at"`
}

type DiscrepancyReportRow struct {
	ID                uuid.UUID  `json:"id"`
	DiscrepancyNumber string     `json:"discrepancy_number"`
	SimakNumber       string     `json:"simak_number"`
	WorkerName        string     `json:"worker_name"`
	ItemName          string     `json:"item_name"`
	DiscrepancyType   string     `json:"discrepancy_type"`
	ExpectedQty       int        `json:"expected_qty"`
	ActualQty         int        `json:"actual_qty"`
	Status            string     `json:"status"`
	ResolutionNotes   string     `json:"resolution_notes"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

type ReportFilterParams struct {
	DateFrom       string `form:"date_from"`
	DateTo         string `form:"date_to"`
	DepartmentID   string `form:"department_id"`
	CompanyID      string `form:"company_id"`
	Direction      string `form:"direction"`
	Status         string `form:"status"`
	GateName       string `form:"gate_name"`
	InspectionType string `form:"inspection_type"`
	Search         string `form:"search"`
	Resolved       string `form:"resolved"` // "true", "false", ""
	Page           int    `form:"page,default=1"`
	PageSize       int    `form:"page_size,default=20"`
}
