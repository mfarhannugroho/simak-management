package approval

import (
	"time"

	"github.com/google/uuid"
	"simak-backend/internal/simak"
	"simak-backend/internal/worker"
)

// ApprovalDecisionRequest represents the payload for approving, rejecting, or requesting revision
type ApprovalDecisionRequest struct {
	Decision string `json:"decision" binding:"required"` // "APPROVE", "REJECT", "REQUEST_REVISION"
	Comments string `json:"comments" binding:"required"`
}

// ApprovalMatrixStepInput represents step definition when creating/updating an approval matrix
type ApprovalMatrixStepInput struct {
	StepOrder      int     `json:"step_order" binding:"required"`
	StepName       string  `json:"step_name" binding:"required"`
	RequiredRole   string  `json:"required_role" binding:"required"`
	SpecificUserID *string `json:"specific_user_id"`
}

// CreateMatrixRequest represents the payload to create a new approval matrix
type CreateMatrixRequest struct {
	Name         string                    `json:"name" binding:"required"`
	Description  string                    `json:"description"`
	DepartmentID *string                   `json:"department_id"`
	CompanyID    *string                   `json:"company_id"`
	IsActive     bool                      `json:"is_active"`
	Steps        []ApprovalMatrixStepInput `json:"steps" binding:"required,min=1"`
}

// UpdateMatrixRequest represents the payload to update an existing approval matrix
type UpdateMatrixRequest struct {
	Name         string                    `json:"name" binding:"required"`
	Description  string                    `json:"description"`
	DepartmentID *string                   `json:"department_id"`
	CompanyID    *string                   `json:"company_id"`
	IsActive     bool                      `json:"is_active"`
	Steps        []ApprovalMatrixStepInput `json:"steps" binding:"required,min=1"`
}

// PendingApprovalQuery represents query parameters for the approver queue
type PendingApprovalQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Search string `form:"search"`
	Status string `form:"status"` // "PENDING", "COMPLETED", "ALL"
}

// PendingApprovalItem represents an application awaiting the reviewer's attention
type PendingApprovalItem struct {
	ApprovalID      uuid.UUID      `json:"approval_id"`
	SimakID         uuid.UUID      `json:"simak_id"`
	SimakNumber     string         `json:"simak_number"`
	Purpose         string         `json:"purpose"`
	Direction       string         `json:"direction"`
	DestinationArea string         `json:"destination_area"`
	StartDate       time.Time      `json:"start_date"`
	EndDate         time.Time      `json:"end_date"`
	SimakStatus     string         `json:"simak_status"`
	CurrentStep     int            `json:"current_step"`
	TotalSteps      int            `json:"total_steps"`
	StepName        string         `json:"step_name"`
	RequiredRole    string         `json:"required_role"`
	StepStatus      string         `json:"step_status"`
	SubmittedAt     time.Time      `json:"submitted_at"`
	TotalItems      int            `json:"total_items"`
	Worker          *worker.Worker `json:"worker,omitempty"`
}

// ApprovalDetailResponse represents the full review detail for an approver
type ApprovalDetailResponse struct {
	Simak           *simak.Simak    `json:"simak"`
	CurrentApproval *SimakApproval  `json:"current_approval,omitempty"`
	WorkflowSteps   []SimakApproval `json:"workflow_steps"`
	CanTakeAction   bool            `json:"can_take_action"`
	RestrictionNote string          `json:"restriction_note,omitempty"`
}
