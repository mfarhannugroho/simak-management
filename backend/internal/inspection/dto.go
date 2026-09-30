package inspection

import "time"

type InspectionItemInput struct {
	SimakItemID  *string `json:"simak_item_id"`
	ItemName     string  `json:"item_name" binding:"required"`
	Brand        string  `json:"brand"`
	SerialNumber string  `json:"serial_number"`
	ExpectedQty  int     `json:"expected_qty"`
	ActualQty    int     `json:"actual_qty"`
	Unit         string  `json:"unit"`
	Status       string  `json:"status"` // "MATCH", "MISMATCH", "UNREGISTERED"
	Notes        string  `json:"notes"`
	PhotoURL     string  `json:"photo_url"`
}

type CreateInspectionRequest struct {
	SimakID        string                `json:"simak_id" binding:"required"`
	GateName       string                `json:"gate_name"`
	InspectionType string                `json:"inspection_type" binding:"required"` // "MASUK", "KELUAR"
	Notes          string                `json:"notes"`
	Items          []InspectionItemInput `json:"items" binding:"required,min=1"`
}

type InspectionFilterParams struct {
	Page           int    `form:"page,default=1"`
	Limit          int    `form:"limit,default=10"`
	SimakID        string `form:"simak_id"`
	GateName       string `form:"gate_name"`
	InspectionType string `form:"inspection_type"`
	Status         string `form:"status"`
	Search         string `form:"search"`
	Date           string `form:"date"`
}

type GateDashboardMetrics struct {
	InspectionsToday int64        `json:"inspections_today"`
	ActiveSimaks     int64        `json:"active_simaks"`
	MismatchCount    int64        `json:"mismatch_count"`
	RecentMovements  []Inspection `json:"recent_movements"`
}

type InspectionListItem struct {
	ID                   string    `json:"id"`
	SimakID              string    `json:"simak_id"`
	SimakNumber          string    `json:"simak_number"`
	InspectorName        string    `json:"inspector_name"`
	WorkerName           string    `json:"worker_name"`
	Company              string    `json:"company"`
	GateName             string    `json:"gate_name"`
	InspectionType       string    `json:"inspection_type"`
	Status               string    `json:"status"`
	TotalRegisteredItems int       `json:"total_registered_items"`
	TotalActualItems     int       `json:"total_actual_items"`
	Notes                string    `json:"notes"`
	InspectedAt          time.Time `json:"inspected_at"`
}
