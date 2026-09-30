package discrepancy

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/inspection"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GenerateDiscrepancyNumber formats a unique case ID: DISC/INKA/YYYY/XXXXXX
func (s *Service) GenerateDiscrepancyNumber(tx *gorm.DB) (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("DISC/INKA/%d/", year)

	var count int64
	err := tx.Model(&Discrepancy{}).
		Where("discrepancy_number LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	seq := count + 1
	return fmt.Sprintf("%s%06d", prefix, seq), nil
}

// AutoCreateFromInspectionInput generates discrepancy cases automatically when an inspection mismatch occurs
func (s *Service) AutoCreateFromInspectionInput(tx *gorm.DB, inspectionID uuid.UUID, simakID uuid.UUID, inspectorID uuid.UUID, items []inspection.InspectionItemInput) error {
	for _, it := range items {
		// Only create discrepancy for items that have discrepancy
		if it.ActualQty == it.ExpectedQty && it.Status != "UNREGISTERED" && it.Status != "MISMATCH" {
			continue
		}

		discType := TypeQuantityMismatch
		if it.Status == "UNREGISTERED" || it.ExpectedQty == 0 {
			discType = TypeUnregisteredItem
		} else if it.ActualQty < it.ExpectedQty {
			discType = TypeMissingItem
		} else if it.ActualQty > it.ExpectedQty {
			discType = TypeQuantityMismatch
		}

		discNum, err := s.GenerateDiscrepancyNumber(tx)
		if err != nil {
			return err
		}

		var inspectionItemUUID *uuid.UUID
		if it.SimakItemID != nil && *it.SimakItemID != "" {
			parsed, err := uuid.Parse(*it.SimakItemID)
			if err == nil {
				inspectionItemUUID = &parsed
			}
		}

		desc := it.Notes
		if desc == "" {
			if discType == TypeUnregisteredItem {
				desc = fmt.Sprintf("Barang temuan fisik tidak terdaftar dalam SIMAK: %s (Qty: %d %s)", it.ItemName, it.ActualQty, it.Unit)
			} else {
				desc = fmt.Sprintf("Selisih fisik: membawa %d %s dari %d %s yang diizinkan", it.ActualQty, it.Unit, it.ExpectedQty, it.Unit)
			}
		}

		disc := Discrepancy{
			DiscrepancyNumber: discNum,
			SimakID:           simakID,
			InspectionID:      inspectionID,
			InspectionItemID:  inspectionItemUUID,
			ItemName:          it.ItemName,
			Brand:             it.Brand,
			SerialNumber:      it.SerialNumber,
			DiscrepancyType:   discType,
			ExpectedQty:       it.ExpectedQty,
			ActualQty:         it.ActualQty,
			Unit:              it.Unit,
			Description:       desc,
			EvidencePhotoURL:  it.PhotoURL,
			ReportedBy:        inspectorID,
			Status:            StatusOpen,
		}

		if err := tx.Create(&disc).Error; err != nil {
			return err
		}
	}

	return nil
}

// CreateDiscrepancy manually creates a discrepancy record
func (s *Service) CreateDiscrepancy(reporterID uuid.UUID, req CreateDiscrepancyRequest) (*Discrepancy, error) {
	simakUUID, err := uuid.Parse(req.SimakID)
	if err != nil {
		return nil, errors.New("ID SIMAK tidak valid")
	}

	insUUID, err := uuid.Parse(req.InspectionID)
	if err != nil {
		return nil, errors.New("ID inspeksi tidak valid")
	}

	var discItemUUID *uuid.UUID
	if req.InspectionItemID != nil && *req.InspectionItemID != "" {
		parsed, err := uuid.Parse(*req.InspectionItemID)
		if err == nil {
			discItemUUID = &parsed
		}
	}

	var created *Discrepancy
	err = s.db.Transaction(func(tx *gorm.DB) error {
		num, err := s.GenerateDiscrepancyNumber(tx)
		if err != nil {
			return err
		}

		unit := req.Unit
		if unit == "" {
			unit = "Unit"
		}

		disc := Discrepancy{
			DiscrepancyNumber: num,
			SimakID:           simakUUID,
			InspectionID:      insUUID,
			InspectionItemID:  discItemUUID,
			ItemName:          req.ItemName,
			Brand:             req.Brand,
			SerialNumber:      req.SerialNumber,
			DiscrepancyType:   req.DiscrepancyType,
			ExpectedQty:       req.ExpectedQty,
			ActualQty:         req.ActualQty,
			Unit:              unit,
			Description:       req.Description,
			EvidencePhotoURL:  req.EvidencePhotoURL,
			ReportedBy:        reporterID,
			Status:            StatusOpen,
		}

		if err := tx.Create(&disc).Error; err != nil {
			return err
		}

		created = &disc
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetDiscrepancyByID(created.ID)
}

// ListDiscrepancies retrieves filtered and paginated discrepancy list
func (s *Service) ListDiscrepancies(p DiscrepancyFilterParams) ([]DiscrepancyListItem, int64, error) {
	query := s.db.Table("discrepancies d").
		Select(`
			d.id,
			d.discrepancy_number,
			d.simak_id,
			s.simak_number,
			w.full_name as worker_name,
			COALESCE(c.name, 'PT INKA (Persero)') as company,
			d.item_name,
			d.brand,
			d.serial_number,
			d.discrepancy_type,
			d.expected_qty,
			d.actual_qty,
			d.unit,
			d.status,
			u.email as reported_by_name,
			d.created_at
		`).
		Joins("JOIN simak s ON s.id = d.simak_id").
		Joins("JOIN workers w ON w.id = s.worker_id").
		Joins("LEFT JOIN companies c ON c.id = w.company_id").
		Joins("LEFT JOIN users u ON u.id = d.reported_by")

	if p.Status != "" && p.Status != "ALL" {
		query = query.Where("d.status = ?", p.Status)
	}

	if p.DiscrepancyType != "" && p.DiscrepancyType != "ALL" {
		query = query.Where("d.discrepancy_type = ?", p.DiscrepancyType)
	}

	if p.SimakID != "" {
		simakUUID, err := uuid.Parse(p.SimakID)
		if err == nil {
			query = query.Where("d.simak_id = ?", simakUUID)
		}
	}

	if p.Date != "" {
		query = query.Where("DATE(d.created_at) = ?", p.Date)
	}

	if p.Search != "" {
		term := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("(LOWER(d.discrepancy_number) LIKE ? OR LOWER(s.simak_number) LIKE ? OR LOWER(d.item_name) LIKE ? OR LOWER(w.full_name) LIKE ? OR LOWER(c.name) LIKE ?)", term, term, term, term, term)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.Limit

	var items []DiscrepancyListItem
	if err := query.Order("d.created_at desc").Offset(offset).Limit(p.Limit).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetDiscrepancyByID loads full discrepancy with relations
func (s *Service) GetDiscrepancyByID(id uuid.UUID) (*Discrepancy, error) {
	var disc Discrepancy
	err := s.db.
		Preload("Simak").
		Preload("Simak.WorkerProfile").
		Preload("Simak.WorkerProfile.Company").
		Preload("Simak.WorkerProfile.Department").
		Preload("Inspection").
		Preload("InspectionItem").
		Preload("Reporter").
		Preload("Reviewer").
		First(&disc, "id = ?", id).Error
	if err != nil {
		return nil, errors.New("Data temuan ketidaksesuaian (discrepancy) tidak ditemukan")
	}
	return &disc, nil
}

// StartReview marks discrepancy status as IN_REVIEW by supervisor
func (s *Service) StartReview(id uuid.UUID, supervisorID uuid.UUID) (*Discrepancy, error) {
	var disc Discrepancy
	if err := s.db.First(&disc, "id = ?", id).Error; err != nil {
		return nil, errors.New("Data ketidaksesuaian tidak ditemukan")
	}

	if disc.Status == StatusResolved || disc.Status == StatusClosed {
		return nil, errors.New("Kasus ketidaksesuaian ini telah diselesaikan sebelumnya")
	}

	now := time.Now()
	disc.Status = StatusInReview
	disc.ReviewedBy = &supervisorID
	disc.ReviewedAt = &now

	if err := s.db.Save(&disc).Error; err != nil {
		return nil, err
	}

	return s.GetDiscrepancyByID(id)
}

// ResolveDiscrepancy resolves or closes a discrepancy investigation
func (s *Service) ResolveDiscrepancy(id uuid.UUID, supervisorID uuid.UUID, req ResolveDiscrepancyRequest) (*Discrepancy, error) {
	var disc Discrepancy
	if err := s.db.First(&disc, "id = ?", id).Error; err != nil {
		return nil, errors.New("Data ketidaksesuaian tidak ditemukan")
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if targetStatus != StatusClosed {
		targetStatus = StatusResolved
	}

	now := time.Now()
	disc.Status = targetStatus
	disc.ResolutionAction = req.ResolutionAction
	disc.ResolutionNotes = req.ResolutionNotes
	disc.ReviewedBy = &supervisorID
	disc.ReviewedAt = &now

	if err := s.db.Save(&disc).Error; err != nil {
		return nil, err
	}

	return s.GetDiscrepancyByID(id)
}

// GetMetrics returns discrepancy KPI counters
func (s *Service) GetMetrics() (*DiscrepancyMetrics, error) {
	var total, open, inReview, resolved, closed int64

	s.db.Model(&Discrepancy{}).Count(&total)
	s.db.Model(&Discrepancy{}).Where("status = ?", StatusOpen).Count(&open)
	s.db.Model(&Discrepancy{}).Where("status = ?", StatusInReview).Count(&inReview)
	s.db.Model(&Discrepancy{}).Where("status = ?", StatusResolved).Count(&resolved)
	s.db.Model(&Discrepancy{}).Where("status = ?", StatusClosed).Count(&closed)

	return &DiscrepancyMetrics{
		Total:    total,
		Open:     open,
		InReview: inReview,
		Resolved: resolved,
		Closed:   closed,
	}, nil
}
