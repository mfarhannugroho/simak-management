package discrepancy

import "time"

type CreateDiscrepancyRequest struct {
	SimakID          string  `json:"simak_id" binding:"required"`
	InspectionID     string  `json:"inspection_id" binding:"required"`
	InspectionItemID *string `json:"inspection_item_id"`
	ItemName         string  `json:"item_name" binding:"required"`
	Brand            string  `json:"brand"`
	SerialNumber     string  `json:"serial_number"`
	DiscrepancyType  string  `json:"discrepancy_type" binding:"required"` // QUANTITY_MISMATCH, UNREGISTERED_ITEM, MISSING_ITEM, ITEM_DESCRIPTION_MISMATCH, OTHER
	ExpectedQty      int     `json:"expected_qty"`
	ActualQty        int     `json:"actual_qty"`
	Unit             string  `json:"unit"`
	Description      string  `json:"description"`
	EvidencePhotoURL string  `json:"evidence_photo_url"`
}

type ResolveDiscrepancyRequest struct {
	ResolutionAction string `json:"resolution_action" binding:"required"` // ALLOW_ENTRY, DENY_ENTRY, CONFISCATE_HOLD_AT_GATE, RETURN_TO_VENDOR, REVISE_PERMIT, WARNING_ISSUED
	ResolutionNotes  string `json:"resolution_notes" binding:"required"`
	Status           string `json:"status"` // RESOLVED, CLOSED
}

type DiscrepancyFilterParams struct {
	Page            int    `form:"page,default=1"`
	Limit           int    `form:"limit,default=10"`
	Status          string `form:"status"`
	DiscrepancyType string `form:"discrepancy_type"`
	SimakID         string `form:"simak_id"`
	Search          string `form:"search"`
	Date            string `form:"date"`
}

type DiscrepancyListItem struct {
	ID                string    `json:"id"`
	DiscrepancyNumber string    `json:"discrepancy_number"`
	SimakID           string    `json:"simak_id"`
	SimakNumber       string    `json:"simak_number"`
	WorkerName        string    `json:"worker_name"`
	Company           string    `json:"company"`
	ItemName          string    `json:"item_name"`
	Brand             string    `json:"brand"`
	SerialNumber      string    `json:"serial_number"`
	DiscrepancyType   string    `json:"discrepancy_type"`
	ExpectedQty       int       `json:"expected_qty"`
	ActualQty         int       `json:"actual_qty"`
	Unit              string    `json:"unit"`
	Status            string    `json:"status"`
	ReportedByName    string    `json:"reported_by_name"`
	CreatedAt         time.Time `json:"created_at"`
}

type DiscrepancyMetrics struct {
	Total    int64 `json:"total"`
	Open     int64 `json:"open"`
	InReview int64 `json:"in_review"`
	Resolved int64 `json:"resolved"`
	Closed   int64 `json:"closed"`
}
