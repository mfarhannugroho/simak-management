package inspection

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/simak"
)

type MismatchHook func(tx *gorm.DB, inspectionID uuid.UUID, simakID uuid.UUID, inspectorID uuid.UUID, items []InspectionItemInput) error

type Service struct {
	db         *gorm.DB
	onMismatch MismatchHook
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SetMismatchHook(hook MismatchHook) {
	s.onMismatch = hook
}

// RecordInspection processes and records physical checklist results conducted by a Gate Officer
func (s *Service) RecordInspection(inspectorID uuid.UUID, req CreateInspectionRequest) (*Inspection, error) {
	simakUUID, err := uuid.Parse(req.SimakID)
	if err != nil {
		return nil, errors.New("ID SIMAK tidak valid")
	}

	var item simak.Simak
	err = s.db.
		Preload("WorkerProfile").
		Preload("Items").
		First(&item, "id = ?", simakUUID).Error
	if err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	allowedStatuses := map[string]bool{
		simak.StatusIssued:    true,
		simak.StatusActive:    true,
		simak.StatusVerified:  true,
		simak.StatusApproved:  true, // allowed in case issued directly
		simak.StatusCompleted: true,
	}

	if !allowedStatuses[item.Status] {
		return nil, fmt.Errorf("SIMAK berstatus %s tidak dapat diverifikasi di pos gerbang", item.Status)
	}

	gateName := req.GateName
	if gateName == "" {
		gateName = "Pos Gerbang Utama PT INKA (Madiun)"
	}

	insType := strings.ToUpper(strings.TrimSpace(req.InspectionType))
	if insType != "MASUK" && insType != "KELUAR" {
		insType = "MASUK"
	}

	// Calculate quantities and determine overall inspection status
	overallStatus := "MATCH"
	totalReg := 0
	totalAct := 0

	for _, it := range req.Items {
		totalReg += it.ExpectedQty
		totalAct += it.ActualQty

		if it.Status == "MISMATCH" || it.Status == "UNREGISTERED" || it.ActualQty != it.ExpectedQty {
			overallStatus = "MISMATCH"
		}
	}

	now := time.Now()

	inspection := Inspection{
		SimakID:              simakUUID,
		InspectorID:          inspectorID,
		GateName:             gateName,
		InspectionType:       insType,
		Status:               overallStatus,
		TotalRegisteredItems: totalReg,
		TotalActualItems:     totalAct,
		Notes:                req.Notes,
		InspectedAt:          now,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&inspection).Error; err != nil {
			return err
		}

		for _, it := range req.Items {
			var simakItemUUID *uuid.UUID
			if it.SimakItemID != nil && *it.SimakItemID != "" {
				parsed, err := uuid.Parse(*it.SimakItemID)
				if err == nil {
					simakItemUUID = &parsed
				}
			}

			itemStatus := it.Status
			if itemStatus == "" {
				if it.ActualQty == it.ExpectedQty {
					itemStatus = "MATCH"
				} else {
					itemStatus = "MISMATCH"
				}
			}

			itemRecord := InspectionItem{
				InspectionID: inspection.ID,
				SimakItemID:  simakItemUUID,
				ItemName:     it.ItemName,
				Brand:        it.Brand,
				SerialNumber: it.SerialNumber,
				ExpectedQty:  it.ExpectedQty,
				ActualQty:    it.ActualQty,
				Unit:         it.Unit,
				Status:       itemStatus,
				Notes:        it.Notes,
				PhotoURL:     it.PhotoURL,
			}
			if err := tx.Create(&itemRecord).Error; err != nil {
				return err
			}
		}

		// Update SIMAK lifecycle status
		if overallStatus == "MATCH" {
			if insType == "KELUAR" && item.Direction != simak.DirectionMasukKeluar {
				item.Status = simak.StatusCompleted
			} else {
				item.Status = simak.StatusVerified
			}
		} else {
			// If mismatch, mark SIMAK as VERIFIED with discrepancy note
			item.Status = simak.StatusVerified
			if req.Notes != "" {
				item.Notes = fmt.Sprintf("%s [Temuan Gate %s: %s]", item.Notes, now.Format("02/01/2006"), req.Notes)
			}

			// Automatically trigger Discrepancy creation
			if s.onMismatch != nil {
				if err := s.onMismatch(tx, inspection.ID, item.ID, inspectorID, req.Items); err != nil {
					return err
				}
			}
		}

		return tx.Save(&item).Error
	})

	if err != nil {
		return nil, err
	}

	return s.GetInspectionByID(inspection.ID)
}

// GetInspectionByID returns full inspection details with line items
func (s *Service) GetInspectionByID(id uuid.UUID) (*Inspection, error) {
	var inspection Inspection
	err := s.db.
		Preload("Simak").
		Preload("Simak.WorkerProfile").
		Preload("Simak.WorkerProfile.Company").
		Preload("Simak.WorkerProfile.Department").
		Preload("Inspector").
		Preload("Items").
		First(&inspection, "id = ?", id).Error
	if err != nil {
		return nil, errors.New("Data hasil pemeriksaan tidak ditemukan")
	}
	return &inspection, nil
}

// ListInspections retrieves filterable gate movement history
func (s *Service) ListInspections(p InspectionFilterParams) ([]InspectionListItem, int64, error) {
	query := s.db.Table("inspections i").
		Select(`
			i.id,
			i.simak_id,
			s.simak_number,
			u.email as inspector_name,
			w.full_name as worker_name,
			COALESCE(c.name, 'PT INKA (Persero)') as company,
			i.gate_name,
			i.inspection_type,
			i.status,
			i.total_registered_items,
			i.total_actual_items,
			i.notes,
			i.inspected_at
		`).
		Joins("JOIN simak s ON s.id = i.simak_id").
		Joins("JOIN workers w ON w.id = s.worker_id").
		Joins("LEFT JOIN companies c ON c.id = w.company_id").
		Joins("LEFT JOIN users u ON u.id = i.inspector_id")

	if p.SimakID != "" {
		simakUUID, err := uuid.Parse(p.SimakID)
		if err == nil {
			query = query.Where("i.simak_id = ?", simakUUID)
		}
	}

	if p.InspectionType != "" && p.InspectionType != "ALL" {
		query = query.Where("i.inspection_type = ?", p.InspectionType)
	}

	if p.Status != "" && p.Status != "ALL" {
		query = query.Where("i.status = ?", p.Status)
	}

	if p.GateName != "" {
		query = query.Where("i.gate_name ILIKE ?", "%"+p.GateName+"%")
	}

	if p.Date != "" {
		query = query.Where("DATE(i.inspected_at) = ?", p.Date)
	}

	if p.Search != "" {
		term := "%" + strings.ToLower(p.Search) + "%"
		query = query.Where("(LOWER(s.simak_number) LIKE ? OR LOWER(w.full_name) LIKE ? OR LOWER(c.name) LIKE ?)", term, term, term)
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

	var items []InspectionListItem
	if err := query.Order("i.inspected_at desc").Offset(offset).Limit(p.Limit).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetGateDashboardMetrics calculates realtime operational counters for gate officers
func (s *Service) GetGateDashboardMetrics() (*GateDashboardMetrics, error) {
	var todayCount int64
	var activeCount int64
	var mismatchCount int64

	todayStr := time.Now().Format("2006-01-02")

	// Inspections today
	s.db.Model(&Inspection{}).Where("DATE(inspected_at) = ?", todayStr).Count(&todayCount)

	// Active SIMAKs
	s.db.Model(&simak.Simak{}).
		Where("status IN ('ISSUED', 'ACTIVE', 'VERIFIED') AND end_date >= CURRENT_DATE").
		Count(&activeCount)

	// Mismatches count
	s.db.Model(&Inspection{}).Where("status = 'MISMATCH'").Count(&mismatchCount)

	// Recent 5 movements
	var recent []Inspection
	s.db.
		Preload("Simak").
		Preload("Simak.WorkerProfile").
		Preload("Simak.WorkerProfile.Company").
		Order("inspected_at desc").
		Limit(5).
		Find(&recent)

	return &GateDashboardMetrics{
		InspectionsToday: todayCount,
		ActiveSimaks:     activeCount,
		MismatchCount:    mismatchCount,
		RecentMovements:  recent,
	}, nil
}
