package approval

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/auth"
	"simak-backend/internal/simak"
	"simak-backend/internal/worker"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GenerateApprovalWorkflow builds approval steps for a submitted SIMAK based on the matching ApprovalMatrix
func (s *Service) GenerateApprovalWorkflow(tx *gorm.DB, simakObj *simak.Simak) error {
	if tx == nil {
		tx = s.db
	}

	// 1. Get worker profile to inspect DepartmentID and CompanyID
	var applicant worker.Worker
	if err := tx.First(&applicant, "id = ?", simakObj.ApplicantID).Error; err != nil {
		return fmt.Errorf("applicant worker profile not found: %w", err)
	}

	// 2. Resolve matching approval matrix
	var matrix ApprovalMatrix
	found := false

	// Check department matrix first
	if applicant.DepartmentID != nil {
		if err := tx.Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order asc")
		}).First(&matrix, "is_active = true AND department_id = ?", *applicant.DepartmentID).Error; err == nil {
			found = true
		}
	}

	// Check company matrix second
	if !found && applicant.CompanyID != nil {
		if err := tx.Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order asc")
		}).First(&matrix, "is_active = true AND company_id = ?", *applicant.CompanyID).Error; err == nil {
			found = true
		}
	}

	// Check global default matrix
	if !found {
		if err := tx.Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order asc")
		}).First(&matrix, "is_active = true AND department_id IS NULL AND company_id IS NULL").Error; err == nil {
			found = true
		}
	}

	// Clean up previous approval entries (e.g. from previous revision submission)
	if err := tx.Where("simak_id = ?", simakObj.ID).Delete(&SimakApproval{}).Error; err != nil {
		return err
	}

	// 3. Instantiate approval steps
	if found && len(matrix.Steps) > 0 {
		for i, st := range matrix.Steps {
			status := "WAITING"
			if i == 0 {
				status = "PENDING"
			}
			approval := SimakApproval{
				SimakID:        simakObj.ID,
				StepOrder:      st.StepOrder,
				StepName:       st.StepName,
				RequiredRole:   st.RequiredRole,
				AssignedUserID: st.SpecificUserID,
				Status:         status,
			}
			if err := tx.Create(&approval).Error; err != nil {
				return err
			}
		}
	} else {
		// Fallback default 2-tier approval if no matrix exists
		defaultSteps := []struct {
			order int
			name  string
			role  string
		}{
			{1, "Pemeriksaan Manajer / PIC Terkait", "Approver"},
			{2, "Verifikasi Supervisor Pengawas Mutu & Area", "Supervisor"},
		}

		for i, st := range defaultSteps {
			status := "WAITING"
			if i == 0 {
				status = "PENDING"
			}
			approval := SimakApproval{
				SimakID:      simakObj.ID,
				StepOrder:    st.order,
				StepName:     st.name,
				RequiredRole: st.role,
				Status:       status,
			}
			if err := tx.Create(&approval).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPendingApprovals retrieves applications requiring review by the user with Segregation of Duties (SoD) enforcement
func (s *Service) GetPendingApprovals(userID uuid.UUID, userRoles []string, q PendingApprovalQuery) ([]PendingApprovalItem, int64, error) {
	isAdmin := containsRole(userRoles, "Admin")

	query := s.db.Table("simak_approvals sa").
		Select(`
			sa.id as approval_id,
			s.id as simak_id,
			s.simak_number,
			s.purpose,
			s.direction,
			s.destination_area,
			s.start_date,
			s.end_date,
			s.status as simak_status,
			sa.step_order as current_step,
			sa.step_name,
			sa.required_role,
			sa.status as step_status,
			s.created_at as submitted_at,
			(SELECT COUNT(*) FROM simak_items si WHERE si.simak_id = s.id) as total_items,
			(SELECT COUNT(*) FROM simak_approvals total_sa WHERE total_sa.simak_id = s.id) as total_steps
		`).
		Joins("JOIN simak s ON s.id = sa.simak_id").
		Joins("JOIN workers w ON w.id = s.worker_id")

	if q.Status == "COMPLETED" || q.Status == "HISTORY" {
		query = query.Where("sa.action_by_user_id = ?", userID)
	} else {
		// Pending approvals for this user
		query = query.Where("sa.status = 'PENDING'").
			Where("s.status IN ('SUBMITTED', 'UNDER_REVIEW')")

		if !isAdmin {
			query = query.Where("sa.required_role IN (?)", userRoles).
				Where("(sa.assigned_user_id IS NULL OR sa.assigned_user_id = ?)", userID)
		}

		// Segregation of Duties (SoD): The applicant cannot approve their own application!
		query = query.Where("(w.user_id IS NULL OR w.user_id != ?)", userID)
	}

	if q.Search != "" {
		searchTerm := "%" + strings.ToLower(q.Search) + "%"
		query = query.Where("(LOWER(s.simak_number) LIKE ? OR LOWER(s.purpose) LIKE ? OR LOWER(w.full_name) LIKE ?)", searchTerm, searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if q.Limit <= 0 {
		q.Limit = 10
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.Limit

	type ResultRow struct {
		ApprovalID      uuid.UUID `json:"approval_id"`
		SimakID         uuid.UUID `json:"simak_id"`
		SimakNumber     *string   `json:"simak_number"`
		Purpose         string    `json:"purpose"`
		Direction       string    `json:"direction"`
		DestinationArea string    `json:"destination_area"`
		StartDate       time.Time `json:"start_date"`
		EndDate         time.Time `json:"end_date"`
		SimakStatus     string    `json:"simak_status"`
		CurrentStep     int       `json:"current_step"`
		StepName        string    `json:"step_name"`
		RequiredRole    string    `json:"required_role"`
		StepStatus      string    `json:"step_status"`
		SubmittedAt     time.Time `json:"submitted_at"`
		TotalItems      int       `json:"total_items"`
		TotalSteps      int       `json:"total_steps"`
	}

	var rows []ResultRow
	if err := query.Order("s.created_at desc").Offset(offset).Limit(q.Limit).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]PendingApprovalItem, len(rows))
	for i, r := range rows {
		simakNum := ""
		if r.SimakNumber != nil {
			simakNum = *r.SimakNumber
		}

		// Fetch worker profile
		var w worker.Worker
		s.db.Preload("Company").Preload("Department").
			Joins("JOIN simak s ON s.worker_id = workers.id").
			Where("s.id = ?", r.SimakID).
			First(&w)

		items[i] = PendingApprovalItem{
			ApprovalID:      r.ApprovalID,
			SimakID:         r.SimakID,
			SimakNumber:     simakNum,
			Purpose:         r.Purpose,
			Direction:       r.Direction,
			DestinationArea: r.DestinationArea,
			StartDate:       r.StartDate,
			EndDate:         r.EndDate,
			SimakStatus:     r.SimakStatus,
			CurrentStep:     r.CurrentStep,
			TotalSteps:      r.TotalSteps,
			StepName:        r.StepName,
			RequiredRole:    r.RequiredRole,
			StepStatus:      r.StepStatus,
			SubmittedAt:     r.SubmittedAt,
			TotalItems:      r.TotalItems,
			Worker:          &w,
		}
	}

	return items, total, nil
}

// GetApprovalDetail returns complete application details, workflow step history, and SoD authorization state
func (s *Service) GetApprovalDetail(simakID uuid.UUID, userID uuid.UUID, userRoles []string) (*ApprovalDetailResponse, error) {
	var simakObj simak.Simak
	err := s.db.
		Preload("WorkerProfile").
		Preload("WorkerProfile.Company").
		Preload("WorkerProfile.Department").
		Preload("Items").
		Preload("Items.Category").
		Preload("Attachments").
		First(&simakObj, "id = ?", simakID).Error
	if err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	// Fetch all approval steps
	var steps []SimakApproval
	if err := s.db.Where("simak_id = ?", simakID).Order("step_order asc").Find(&steps).Error; err != nil {
		return nil, err
	}

	// Populate action_by_name for each step
	for i := range steps {
		if steps[i].ActionByUserID != nil {
			var reviewer worker.Worker
			if err := s.db.Select("full_name").First(&reviewer, "user_id = ?", *steps[i].ActionByUserID).Error; err == nil && reviewer.FullName != "" {
				steps[i].ActionByName = reviewer.FullName
			} else {
				var u auth.User
				if err := s.db.Select("email").First(&u, "id = ?", *steps[i].ActionByUserID).Error; err == nil {
					steps[i].ActionByName = u.Email
				}
			}
		}
	}

	// Find the active PENDING step
	var currentPending *SimakApproval
	for i := range steps {
		if steps[i].Status == "PENDING" {
			currentPending = &steps[i]
			break
		}
	}

	isAdmin := containsRole(userRoles, "Admin")
	canTakeAction := false
	restrictionNote := ""

	// Check Segregation of Duties (SoD)
	if simakObj.WorkerProfile != nil && simakObj.WorkerProfile.UserID == userID {
		canTakeAction = false
		restrictionNote = "Prinsip Segregation of Duties (SoD): Anda adalah pemohon dari SIMAK ini sehingga tidak diperkenankan memberikan persetujuan atau keputusan."
	} else if currentPending != nil {
		// Check role requirement
		hasRole := containsRole(userRoles, currentPending.RequiredRole) || isAdmin
		isAssigned := currentPending.AssignedUserID == nil || *currentPending.AssignedUserID == userID || isAdmin

		if hasRole && isAssigned {
			canTakeAction = true
		} else {
			restrictionNote = fmt.Sprintf("Langkah ini membutuhkan hak akses role '%s'", currentPending.RequiredRole)
		}
	} else {
		restrictionNote = fmt.Sprintf("Tidak ada langkah persetujuan yang aktif menunggu saat ini (Status SIMAK: %s)", simakObj.Status)
	}

	return &ApprovalDetailResponse{
		Simak:           &simakObj,
		CurrentApproval: currentPending,
		WorkflowSteps:   steps,
		CanTakeAction:   canTakeAction,
		RestrictionNote: restrictionNote,
	}, nil
}

// ProcessDecision handles approving, rejecting, or requesting revision on a SIMAK with strict SoD checking
func (s *Service) ProcessDecision(simakID uuid.UUID, userID uuid.UUID, userRoles []string, req ApprovalDecisionRequest) (*simak.Simak, error) {
	var simakObj simak.Simak
	if err := s.db.Preload("WorkerProfile").First(&simakObj, "id = ?", simakID).Error; err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	// 1. Segregation of Duties (SoD) check
	if simakObj.WorkerProfile != nil && simakObj.WorkerProfile.UserID == userID {
		return nil, errors.New("Prinsip Segregation of Duties (SoD): Anda tidak diizinkan menyetujui atau memproses permohonan yang Anda ajukan sendiri")
	}

	// 2. Find currently active pending step
	var currentStep SimakApproval
	if err := s.db.Where("simak_id = ? AND status = 'PENDING'", simakID).Order("step_order asc").First(&currentStep).Error; err != nil {
		return nil, errors.New("Tidak ada langkah persetujuan yang sedang menunggu keputusan untuk SIMAK ini")
	}

	// 3. Verify user has required role or is Admin
	isAdmin := containsRole(userRoles, "Admin")
	if !isAdmin && !containsRole(userRoles, currentStep.RequiredRole) {
		return nil, fmt.Errorf("Anda tidak memiliki role yang diperlukan (%s) untuk memproses langkah ini", currentStep.RequiredRole)
	}
	if currentStep.AssignedUserID != nil && *currentStep.AssignedUserID != userID && !isAdmin {
		return nil, errors.New("Langkah persetujuan ini ditugaskan khusus untuk pengguna lain")
	}

	now := time.Now()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		decision := strings.ToUpper(strings.TrimSpace(req.Decision))

		switch decision {
		case "APPROVE":
			// Update current step to APPROVED
			currentStep.Status = "APPROVED"
			currentStep.ActionByUserID = &userID
			currentStep.ActionAt = &now
			currentStep.Comments = req.Comments
			if err := tx.Save(&currentStep).Error; err != nil {
				return err
			}

			// Check for next step
			var nextStep SimakApproval
			err := tx.Where("simak_id = ? AND step_order > ?", simakID, currentStep.StepOrder).
				Order("step_order asc").First(&nextStep).Error

			if err == nil {
				// Advance next step to PENDING
				nextStep.Status = "PENDING"
				if err := tx.Save(&nextStep).Error; err != nil {
					return err
				}
				simakObj.Status = simak.StatusUnderReview
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// Final step approved -> SIMAK is APPROVED!
				simakObj.Status = simak.StatusApproved

				// Generate official issued SIMAK reference number: SMK-YYYY-MM-XXXX
				if simakObj.SimakNumber == nil || !strings.HasPrefix(*simakObj.SimakNumber, "SMK-") {
					yearMonth := now.Format("2006-01")
					var count int64
					tx.Model(&simak.Simak{}).Where("simak_number LIKE ?", "SMK-"+yearMonth+"-%").Count(&count)
					officialNum := fmt.Sprintf("SMK-%s-%04d", yearMonth, count+1)
					simakObj.SimakNumber = &officialNum
				}
			} else {
				return err
			}

		case "REQUEST_REVISION":
			currentStep.Status = "REVISION_REQUESTED"
			currentStep.ActionByUserID = &userID
			currentStep.ActionAt = &now
			currentStep.Comments = req.Comments
			if err := tx.Save(&currentStep).Error; err != nil {
				return err
			}

			simakObj.Status = simak.StatusRequiresRevision
			simakObj.RevisionNotes = req.Comments

		case "REJECT":
			currentStep.Status = "REJECTED"
			currentStep.ActionByUserID = &userID
			currentStep.ActionAt = &now
			currentStep.Comments = req.Comments
			if err := tx.Save(&currentStep).Error; err != nil {
				return err
			}

			simakObj.Status = simak.StatusRejected
			simakObj.RejectionReason = req.Comments

		default:
			return fmt.Errorf("Keputusan tidak valid: %s (Gunakan: APPROVE, REJECT, atau REQUEST_REVISION)", req.Decision)
		}

		return tx.Save(&simakObj).Error
	})

	if err != nil {
		return nil, err
	}

	return &simakObj, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Admin Approval Matrix CRUD
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) ListMatrices() ([]ApprovalMatrix, error) {
	var matrices []ApprovalMatrix
	err := s.db.
		Preload("Department").
		Preload("Company").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order asc")
		}).
		Order("created_at desc").
		Find(&matrices).Error
	return matrices, err
}

func (s *Service) GetMatrixByID(id uuid.UUID) (*ApprovalMatrix, error) {
	var matrix ApprovalMatrix
	err := s.db.
		Preload("Department").
		Preload("Company").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order asc")
		}).
		First(&matrix, "id = ?", id).Error
	if err != nil {
		return nil, errors.New("Matriks persetujuan tidak ditemukan")
	}
	return &matrix, nil
}

func (s *Service) CreateMatrix(req CreateMatrixRequest) (*ApprovalMatrix, error) {
	var deptUUID, compUUID *uuid.UUID
	if req.DepartmentID != nil && *req.DepartmentID != "" {
		id, err := uuid.Parse(*req.DepartmentID)
		if err == nil {
			deptUUID = &id
		}
	}
	if req.CompanyID != nil && *req.CompanyID != "" {
		id, err := uuid.Parse(*req.CompanyID)
		if err == nil {
			compUUID = &id
		}
	}

	matrix := ApprovalMatrix{
		Name:         req.Name,
		Description:  req.Description,
		DepartmentID: deptUUID,
		CompanyID:    compUUID,
		IsActive:     req.IsActive,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&matrix).Error; err != nil {
			return err
		}

		for _, st := range req.Steps {
			var userUUID *uuid.UUID
			if st.SpecificUserID != nil && *st.SpecificUserID != "" {
				uID, err := uuid.Parse(*st.SpecificUserID)
				if err == nil {
					userUUID = &uID
				}
			}

			stepRecord := ApprovalMatrixStep{
				MatrixID:       matrix.ID,
				StepOrder:      st.StepOrder,
				StepName:       st.StepName,
				RequiredRole:   st.RequiredRole,
				SpecificUserID: userUUID,
			}
			if err := tx.Create(&stepRecord).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetMatrixByID(matrix.ID)
}

func (s *Service) UpdateMatrix(id uuid.UUID, req UpdateMatrixRequest) (*ApprovalMatrix, error) {
	var matrix ApprovalMatrix
	if err := s.db.First(&matrix, "id = ?", id).Error; err != nil {
		return nil, errors.New("Matriks persetujuan tidak ditemukan")
	}

	var deptUUID, compUUID *uuid.UUID
	if req.DepartmentID != nil && *req.DepartmentID != "" {
		parsed, err := uuid.Parse(*req.DepartmentID)
		if err == nil {
			deptUUID = &parsed
		}
	}
	if req.CompanyID != nil && *req.CompanyID != "" {
		parsed, err := uuid.Parse(*req.CompanyID)
		if err == nil {
			compUUID = &parsed
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		matrix.Name = req.Name
		matrix.Description = req.Description
		matrix.DepartmentID = deptUUID
		matrix.CompanyID = compUUID
		matrix.IsActive = req.IsActive

		if err := tx.Save(&matrix).Error; err != nil {
			return err
		}

		// Replace steps
		if err := tx.Where("matrix_id = ?", id).Delete(&ApprovalMatrixStep{}).Error; err != nil {
			return err
		}

		for _, st := range req.Steps {
			var userUUID *uuid.UUID
			if st.SpecificUserID != nil && *st.SpecificUserID != "" {
				uID, err := uuid.Parse(*st.SpecificUserID)
				if err == nil {
					userUUID = &uID
				}
			}

			stepRecord := ApprovalMatrixStep{
				MatrixID:       id,
				StepOrder:      st.StepOrder,
				StepName:       st.StepName,
				RequiredRole:   st.RequiredRole,
				SpecificUserID: userUUID,
			}
			if err := tx.Create(&stepRecord).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetMatrixByID(id)
}

func (s *Service) DeleteMatrix(id uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("matrix_id = ?", id).Delete(&ApprovalMatrixStep{}).Error; err != nil {
			return err
		}
		return tx.Delete(&ApprovalMatrix{}, "id = ?", id).Error
	})
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}
