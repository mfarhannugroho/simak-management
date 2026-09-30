package dashboard

import "time"

type RoleMetric struct {
	Label       string `json:"label"`
	Value       int64  `json:"value"`
	Description string `json:"description"`
	Trend       string `json:"trend"`
	TrendType   string `json:"trend_type"` // positive, negative, neutral
	Highlight   bool   `json:"highlight"`
}

type DashboardSummaryResponse struct {
	Role        string       `json:"role"`
	UserName    string       `json:"user_name"`
	WorkerID    string       `json:"worker_id"`
	Position    string       `json:"position"`
	Department  string       `json:"department"`
	Company     string       `json:"company"`
	Metrics     []RoleMetric `json:"metrics"`
	TotalSimak  int64        `json:"total_simak"`
	ActiveSimak int64        `json:"active_simak"`
}

type DailyVolumeItem struct {
	Day      string `json:"day"`      // "Sen", "Sel", "Rab", etc.
	Date     string `json:"date"`     // "2026-09-29"
	Masuk    int64  `json:"masuk"`    // Inflow
	Keluar   int64  `json:"keluar"`   // Outflow
	Total    int64  `json:"total"`    // Total movements
	IsPeak   bool   `json:"is_peak"`  // True if maximum day
}

type VolumeTrendResponse struct {
	Period      string            `json:"period"`
	TotalVolume int64             `json:"total_volume"`
	PeakDay     string            `json:"peak_day"`
	Daily       []DailyVolumeItem `json:"daily"`
}

type ActivityItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // INSPECTION, APPROVAL, SIMAK_SUBMITTED, DISCREPANCY
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Variant     string    `json:"variant"` // success, danger, warning, neutral, inka
	Timestamp   time.Time `json:"timestamp"`
	ActionURL   string    `json:"action_url"`
}

type ActivitiesResponse struct {
	Period     string         `json:"period"`
	Activities []ActivityItem `json:"activities"`
}
