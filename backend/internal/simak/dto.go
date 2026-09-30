package simak

// SimakItemInput represents payload for creating/updating a carried tool item
type SimakItemInput struct {
	CategoryID   *string `json:"category_id"`
	ItemName     string  `json:"item_name" binding:"required"`
	Brand        string  `json:"brand"`
	Model        string  `json:"model"`
	SerialNumber string  `json:"serial_number"`
	SizeSpec     string  `json:"size_spec"`
	Quantity     int     `json:"quantity" binding:"required,min=1"`
	Unit         string  `json:"unit" binding:"required"`
	Description  string  `json:"description"`
	PhotoURL     string  `json:"photo_url"`
}

// SimakAttachmentInput represents payload for registering an uploaded file
type SimakAttachmentInput struct {
	FileName       string `json:"file_name" binding:"required"`
	FileURL        string `json:"file_url" binding:"required"`
	FileType       string `json:"file_type"`
	FileSize       int64  `json:"file_size"`
	AttachmentType string `json:"attachment_type"`
}

// CreateSimakRequest represents payload to create a new SIMAK application
type CreateSimakRequest struct {
	WorkerID        string                 `json:"worker_id"`
	Purpose         string                 `json:"purpose" binding:"required"`
	Direction       string                 `json:"direction" binding:"required"`
	DestinationArea string                 `json:"destination_area" binding:"required"`
	StartDate       string                 `json:"start_date" binding:"required"`
	EndDate         string                 `json:"end_date" binding:"required"`
	Notes           string                 `json:"notes"`
	SubmitDirectly  bool                   `json:"submit_directly"`
	Items           []SimakItemInput       `json:"items"`
	Attachments     []SimakAttachmentInput `json:"attachments"`
}

// UpdateSimakRequest represents payload to edit a draft or revised SIMAK
type UpdateSimakRequest struct {
	Purpose         string                 `json:"purpose" binding:"required"`
	Direction       string                 `json:"direction" binding:"required"`
	DestinationArea string                 `json:"destination_area" binding:"required"`
	StartDate       string                 `json:"start_date" binding:"required"`
	EndDate         string                 `json:"end_date" binding:"required"`
	Notes           string                 `json:"notes"`
	Items           []SimakItemInput       `json:"items"`
	Attachments     []SimakAttachmentInput `json:"attachments"`
}

// SubmitSimakRequest optional body for submitting with optional note
type SubmitSimakRequest struct {
	Notes string `json:"notes"`
}

// CancelSimakRequest body for cancelling
type CancelSimakRequest struct {
	Reason string `json:"reason"`
}

// SimakFilterParams for querying permits
type SimakFilterParams struct {
	Status    string
	Direction string
	Search    string
	StartDate string
	EndDate   string
	WorkerID  string
	CompanyID string
	Page      int
	Limit     int
}
