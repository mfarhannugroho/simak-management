package simak

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowGenerator interface {
	GenerateApprovalWorkflow(tx *gorm.DB, simakObj *Simak) error
}

type Service struct {
	db                *gorm.DB
	workflowGenerator WorkflowGenerator
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SetWorkflowGenerator(wg WorkflowGenerator) {
	s.workflowGenerator = wg
}

// GenerateReferenceNumber generates a human-readable draft or request identifier
func (s *Service) GenerateReferenceNumber(prefix string) string {
	now := time.Now()
	randomSuffix := strings.ToUpper(uuid.New().String()[:5])
	return fmt.Sprintf("%s-%s-%s", prefix, now.Format("20060102"), randomSuffix)
}

// Create creates a new SIMAK application with items and attachments
func (s *Service) Create(workerID uuid.UUID, req CreateSimakRequest) (*Simak, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("Format tanggal mulai tidak valid (harus YYYY-MM-DD)")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("Format tanggal selesai tidak valid (harus YYYY-MM-DD)")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("Tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	}

	status := StatusDraft
	numberPrefix := "DFT"
	if req.SubmitDirectly {
		if len(req.Items) == 0 {
			return nil, errors.New("Permohonan SIMAK wajib menyertakan minimal 1 (satu) barang bawaan")
		}
		status = StatusSubmitted
		numberPrefix = "REQ"
	}

	simakNumber := s.GenerateReferenceNumber(numberPrefix)

	simakObj := Simak{
		ApplicantID:     workerID,
		SimakNumber:     &simakNumber,
		Purpose:         req.Purpose,
		Direction:       req.Direction,
		DestinationArea: req.DestinationArea,
		StartDate:       startDate,
		EndDate:         endDate,
		Status:          status,
		Notes:           req.Notes,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&simakObj).Error; err != nil {
			return err
		}

		// Insert Items
		for _, it := range req.Items {
			itemRecord := SimakItem{
				SimakID:      simakObj.ID,
				ItemName:     it.ItemName,
				Brand:        it.Brand,
				Model:        it.Model,
				SerialNumber: it.SerialNumber,
				SizeSpec:     it.SizeSpec,
				Quantity:     it.Quantity,
				Unit:         it.Unit,
				Description:  it.Description,
				PhotoURL:     it.PhotoURL,
			}
			if it.CategoryID != nil && *it.CategoryID != "" {
				catUUID, err := uuid.Parse(*it.CategoryID)
				if err == nil {
					itemRecord.CategoryID = &catUUID
				}
			}
			if err := tx.Create(&itemRecord).Error; err != nil {
				return err
			}
		}

		// Insert Attachments
		for _, att := range req.Attachments {
			attType := att.AttachmentType
			if attType == "" {
				attType = "SURAT_TUGAS"
			}
			attRecord := SimakAttachment{
				SimakID:        simakObj.ID,
				FileName:       att.FileName,
				FileURL:        att.FileURL,
				FileType:       att.FileType,
				FileSize:       att.FileSize,
				AttachmentType: attType,
			}
			if err := tx.Create(&attRecord).Error; err != nil {
				return err
			}
		}

		if status == StatusSubmitted && s.workflowGenerator != nil {
			if err := s.workflowGenerator.GenerateApprovalWorkflow(tx, &simakObj); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetByID(simakObj.ID)
}

// List returns paginated SIMAK applications matching filter parameters
func (s *Service) List(params SimakFilterParams) ([]Simak, int64, error) {
	query := s.db.Model(&Simak{})

	if params.WorkerID != "" {
		workerUUID, err := uuid.Parse(params.WorkerID)
		if err == nil {
			query = query.Where("worker_id = ?", workerUUID)
		}
	}

	if params.Status != "" && params.Status != "ALL" {
		query = query.Where("status = ?", params.Status)
	}

	if params.Direction != "" && params.Direction != "ALL" {
		query = query.Where("direction = ?", params.Direction)
	}

	if params.StartDate != "" {
		query = query.Where("start_date >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("end_date <= ?", params.EndDate)
	}

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Joins("LEFT JOIN workers ON workers.id = simak.worker_id").
			Where("simak.simak_number ILIKE ? OR simak.purpose ILIKE ? OR simak.destination_area ILIKE ? OR workers.full_name ILIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	limit := params.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var list []Simak
	err := query.
		Preload("WorkerProfile").
		Preload("WorkerProfile.Company").
		Preload("WorkerProfile.Department").
		Preload("Items").
		Preload("Items.Category").
		Preload("Attachments").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

// GetByID retrieves a single SIMAK record with all its relational data
func (s *Service) GetByID(id uuid.UUID) (*Simak, error) {
	var item Simak
	err := s.db.
		Preload("WorkerProfile").
		Preload("WorkerProfile.Company").
		Preload("WorkerProfile.Department").
		Preload("Items").
		Preload("Items.Category").
		Preload("Attachments").
		First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update modifies an existing SIMAK if it is in DRAFT or REQUIRES_REVISION status
func (s *Service) Update(id uuid.UUID, workerID uuid.UUID, isAdmin bool, req UpdateSimakRequest) (*Simak, error) {
	var item Simak
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	if !isAdmin && item.ApplicantID != workerID {
		return nil, errors.New("Anda tidak memiliki hak akses untuk mengubah pengajuan ini")
	}

	if item.Status != StatusDraft && item.Status != StatusRequiresRevision {
		return nil, fmt.Errorf("Pengajuan berstatus %s tidak dapat diubah lagi", item.Status)
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("Format tanggal mulai tidak valid (harus YYYY-MM-DD)")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("Format tanggal selesai tidak valid (harus YYYY-MM-DD)")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("Tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update main details
		item.Purpose = req.Purpose
		item.Direction = req.Direction
		item.DestinationArea = req.DestinationArea
		item.StartDate = startDate
		item.EndDate = endDate
		item.Notes = req.Notes

		if err := tx.Save(&item).Error; err != nil {
			return err
		}

		// Replace items
		if err := tx.Where("simak_id = ?", id).Delete(&SimakItem{}).Error; err != nil {
			return err
		}

		for _, it := range req.Items {
			itemRecord := SimakItem{
				SimakID:      id,
				ItemName:     it.ItemName,
				Brand:        it.Brand,
				Model:        it.Model,
				SerialNumber: it.SerialNumber,
				SizeSpec:     it.SizeSpec,
				Quantity:     it.Quantity,
				Unit:         it.Unit,
				Description:  it.Description,
				PhotoURL:     it.PhotoURL,
			}
			if it.CategoryID != nil && *it.CategoryID != "" {
				catUUID, err := uuid.Parse(*it.CategoryID)
				if err == nil {
					itemRecord.CategoryID = &catUUID
				}
			}
			if err := tx.Create(&itemRecord).Error; err != nil {
				return err
			}
		}

		// Replace attachments if provided
		if len(req.Attachments) > 0 {
			if err := tx.Where("simak_id = ?", id).Delete(&SimakAttachment{}).Error; err != nil {
				return err
			}
			for _, att := range req.Attachments {
				attType := att.AttachmentType
				if attType == "" {
					attType = "SURAT_TUGAS"
				}
				attRecord := SimakAttachment{
					SimakID:        id,
					FileName:       att.FileName,
					FileURL:        att.FileURL,
					FileType:       att.FileType,
					FileSize:       att.FileSize,
					AttachmentType: attType,
				}
				if err := tx.Create(&attRecord).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Delete permanently removes a draft SIMAK
func (s *Service) Delete(id uuid.UUID, workerID uuid.UUID, isAdmin bool) error {
	var item Simak
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return errors.New("SIMAK tidak ditemukan")
	}

	if !isAdmin && item.ApplicantID != workerID {
		return errors.New("Anda tidak memiliki hak akses untuk menghapus pengajuan ini")
	}

	if item.Status != StatusDraft {
		return fmt.Errorf("Hanya pengajuan berstatus DRAFT yang dapat dihapus. Status saat ini: %s", item.Status)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("simak_id = ?", id).Delete(&SimakItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("simak_id = ?", id).Delete(&SimakAttachment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&item).Error
	})
}

// Submit transitions SIMAK from DRAFT or REQUIRES_REVISION to SUBMITTED
func (s *Service) Submit(id uuid.UUID, workerID uuid.UUID, isAdmin bool, notes string) (*Simak, error) {
	var item Simak
	if err := s.db.Preload("Items").First(&item, "id = ?", id).Error; err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	if !isAdmin && item.ApplicantID != workerID {
		return nil, errors.New("Anda tidak berhak mengajukan permohonan ini")
	}

	if item.Status != StatusDraft && item.Status != StatusRequiresRevision {
		return nil, fmt.Errorf("SIMAK tidak dapat diajukan karena berstatus %s", item.Status)
	}

	if len(item.Items) == 0 {
		return nil, errors.New("Permohonan SIMAK wajib memiliki minimal 1 (satu) barang bawaan")
	}

	item.Status = StatusSubmitted
	if notes != "" {
		item.Notes = notes
	}

	// Update SIMAK number to REQ- if it was draft
	if item.SimakNumber != nil && strings.HasPrefix(*item.SimakNumber, "DFT-") {
		newNum := s.GenerateReferenceNumber("REQ")
		item.SimakNumber = &newNum
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		if s.workflowGenerator != nil {
			if err := s.workflowGenerator.GenerateApprovalWorkflow(tx, &item); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Cancel transitions a pending SIMAK to CANCELLED
func (s *Service) Cancel(id uuid.UUID, workerID uuid.UUID, isAdmin bool, reason string) (*Simak, error) {
	var item Simak
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	if !isAdmin && item.ApplicantID != workerID {
		return nil, errors.New("Anda tidak berhak membatalkan permohonan ini")
	}

	allowedStatuses := map[string]bool{
		StatusDraft:            true,
		StatusSubmitted:        true,
		StatusUnderReview:      true,
		StatusRequiresRevision: true,
	}

	if !allowedStatuses[item.Status] {
		return nil, fmt.Errorf("SIMAK berstatus %s tidak dapat dibatalkan secara langsung", item.Status)
	}

	item.Status = StatusCancelled
	if reason != "" {
		item.Notes = fmt.Sprintf("%s [Alasan Pembatalan: %s]", item.Notes, reason)
	}

	if err := s.db.Save(&item).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// AddAttachment adds an attachment document to SIMAK
func (s *Service) AddAttachment(simakID uuid.UUID, input SimakAttachmentInput) (*SimakAttachment, error) {
	attType := input.AttachmentType
	if attType == "" {
		attType = "SURAT_TUGAS"
	}
	att := SimakAttachment{
		SimakID:        simakID,
		FileName:       input.FileName,
		FileURL:        input.FileURL,
		FileType:       input.FileType,
		FileSize:       input.FileSize,
		AttachmentType: attType,
	}
	if err := s.db.Create(&att).Error; err != nil {
		return nil, err
	}
	return &att, nil
}

// DeleteAttachment removes an attachment record
func (s *Service) DeleteAttachment(attachmentID uuid.UUID) error {
	return s.db.Delete(&SimakAttachment{}, "id = ?", attachmentID).Error
}
