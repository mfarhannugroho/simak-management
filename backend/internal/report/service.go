package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetSimakReport returns filtered list of SIMAK records
func (s *Service) GetSimakReport(params ReportFilterParams) ([]SimakReportRow, int64, error) {
	query := s.db.Table("simak").
		Select(`
			simak.id,
			COALESCE(simak.simak_number, 'DRAFT') as simak_number,
			COALESCE(workers.full_name, '-') as worker_name,
			COALESCE(workers.worker_id, '-') as worker_nip,
			COALESCE(departments.name, '-') as department_name,
			COALESCE(companies.name, '-') as company_name,
			simak.direction,
			simak.destination_area,
			TO_CHAR(simak.start_date, 'YYYY-MM-DD') as start_date,
			TO_CHAR(simak.end_date, 'YYYY-MM-DD') as end_date,
			simak.status,
			(SELECT COUNT(*) FROM simak_items WHERE simak_items.simak_id = simak.id) as total_items,
			simak.issued_at,
			simak.created_at
		`).
		Joins("LEFT JOIN workers ON workers.id = simak.worker_id").
		Joins("LEFT JOIN departments ON departments.id = workers.department_id").
		Joins("LEFT JOIN companies ON companies.id = workers.company_id")

	if params.DateFrom != "" {
		query = query.Where("simak.start_date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("simak.end_date <= ?", params.DateTo)
	}
	if params.DepartmentID != "" {
		query = query.Where("workers.department_id = ?", params.DepartmentID)
	}
	if params.CompanyID != "" {
		query = query.Where("workers.company_id = ?", params.CompanyID)
	}
	if params.Direction != "" {
		query = query.Where("simak.direction = ?", params.Direction)
	}
	if params.Status != "" {
		query = query.Where("simak.status = ?", params.Status)
	}
	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where("LOWER(simak.simak_number) LIKE ? OR LOWER(workers.full_name) LIKE ? OR LOWER(workers.worker_id) LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := params.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var rows []SimakReportRow
	err := query.Order("simak.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// GetVerificationReport returns gate security inspection audit records
func (s *Service) GetVerificationReport(params ReportFilterParams) ([]VerificationReportRow, int64, error) {
	query := s.db.Table("inspections").
		Select(`
			inspections.id,
			COALESCE(simak.simak_number, 'N/A') as simak_number,
			COALESCE(workers.full_name, '-') as worker_name,
			COALESCE(inspector_workers.full_name, users.email) as inspector_name,
			inspections.gate_name,
			inspections.inspection_type,
			inspections.status,
			inspections.total_registered_items,
			inspections.total_actual_items,
			COALESCE(inspections.notes, '') as notes,
			inspections.inspected_at
		`).
		Joins("LEFT JOIN simak ON simak.id = inspections.simak_id").
		Joins("LEFT JOIN workers ON workers.id = simak.worker_id").
		Joins("LEFT JOIN users ON users.id = inspections.inspector_id").
		Joins("LEFT JOIN workers as inspector_workers ON inspector_workers.user_id = users.id")

	if params.DateFrom != "" {
		query = query.Where("inspections.inspected_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("inspections.inspected_at <= ?", params.DateTo)
	}
	if params.GateName != "" {
		query = query.Where("inspections.gate_name = ?", params.GateName)
	}
	if params.InspectionType != "" {
		query = query.Where("inspections.inspection_type = ?", params.InspectionType)
	}
	if params.Status != "" {
		query = query.Where("inspections.status = ?", params.Status)
	}
	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where("LOWER(simak.simak_number) LIKE ? OR LOWER(workers.full_name) LIKE ? OR LOWER(inspector_workers.full_name) LIKE ? OR LOWER(users.email) LIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := params.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var rows []VerificationReportRow
	err := query.Order("inspections.inspected_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// GetDiscrepancyReport returns discrepancy case history
func (s *Service) GetDiscrepancyReport(params ReportFilterParams) ([]DiscrepancyReportRow, int64, error) {
	query := s.db.Table("discrepancies").
		Select(`
			discrepancies.id,
			discrepancies.discrepancy_number,
			COALESCE(simak.simak_number, 'N/A') as simak_number,
			COALESCE(workers.full_name, '-') as worker_name,
			discrepancies.item_name,
			discrepancies.discrepancy_type,
			discrepancies.expected_qty,
			discrepancies.actual_qty,
			discrepancies.status,
			COALESCE(discrepancies.resolution_notes, '') as resolution_notes,
			discrepancies.reviewed_at,
			discrepancies.created_at
		`).
		Joins("LEFT JOIN inspections ON inspections.id = discrepancies.inspection_id").
		Joins("LEFT JOIN simak ON simak.id = discrepancies.simak_id").
		Joins("LEFT JOIN workers ON workers.id = simak.worker_id")

	if params.DateFrom != "" {
		query = query.Where("discrepancies.created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("discrepancies.created_at <= ?", params.DateTo)
	}
	if params.Status != "" {
		query = query.Where("discrepancies.status = ?", params.Status)
	}
	if params.Resolved == "true" {
		query = query.Where("discrepancies.status IN ('RESOLVED', 'CLOSED')")
	} else if params.Resolved == "false" {
		query = query.Where("discrepancies.status IN ('OPEN', 'IN_REVIEW')")
	}
	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where("LOWER(discrepancies.discrepancy_number) LIKE ? OR LOWER(discrepancies.item_name) LIKE ? OR LOWER(workers.full_name) LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := params.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var rows []DiscrepancyReportRow
	err := query.Order("discrepancies.created_at DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// ExportSimakCSV produces UTF-8 BOM CSV bytes for SIMAK report
func (s *Service) ExportSimakCSV(params ReportFilterParams) ([]byte, error) {
	// Set large limit for full export
	params.Page = 1
	params.PageSize = 10000
	rows, _, err := s.GetSimakReport(params)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	// Write UTF-8 BOM
	buf.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(&buf)
	// Header row
	header := []string{
		"Nomor SIMAK", "Nama Pemohon", "NIP", "Departemen", "Perusahaan",
		"Arah Gerakan", "Area Tujuan", "Tgl Mulai", "Tgl Selesai",
		"Status", "Jumlah Alat", "Tgl Diterbitkan", "Tgl Dibuat",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, r := range rows {
		issuedStr := "-"
		if r.IssuedAt != nil {
			issuedStr = r.IssuedAt.Format("2006-01-02 15:04:05")
		}

		record := []string{
			r.SimakNumber,
			r.WorkerName,
			r.WorkerNIP,
			r.DepartmentName,
			r.CompanyName,
			r.Direction,
			r.DestinationArea,
			r.StartDate,
			r.EndDate,
			r.Status,
			fmt.Sprintf("%d", r.TotalItems),
			issuedStr,
			r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportVerificationCSV produces UTF-8 BOM CSV bytes for Inspection report
func (s *Service) ExportVerificationCSV(params ReportFilterParams) ([]byte, error) {
	params.Page = 1
	params.PageSize = 10000
	rows, _, err := s.GetVerificationReport(params)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(&buf)
	header := []string{
		"Nomor SIMAK", "Nama Personil", "Petugas Pemeriksa", "Pos Gerbang",
		"Jenis Inspeksi", "Hasil Status", "Item Terdaftar", "Item Fisik",
		"Catatan Pemeriksaan", "Waktu Verifikasi",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, r := range rows {
		record := []string{
			r.SimakNumber,
			r.WorkerName,
			r.InspectorName,
			r.GateName,
			r.InspectionType,
			r.Status,
			fmt.Sprintf("%d", r.TotalRegisteredItems),
			fmt.Sprintf("%d", r.TotalActualItems),
			r.Notes,
			r.InspectedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ExportDiscrepancyCSV produces UTF-8 BOM CSV bytes for Discrepancy report
func (s *Service) ExportDiscrepancyCSV(params ReportFilterParams) ([]byte, error) {
	params.Page = 1
	params.PageSize = 10000
	rows, _, err := s.GetDiscrepancyReport(params)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(&buf)
	header := []string{
		"Nomor Kasus", "Nomor SIMAK", "Nama Personil", "Nama Peralatan",
		"Tipe Ketidaksesuaian", "Qty Terdaftar", "Qty Riil", "Status Solusi",
		"Catatan Penyelesaian", "Tgl Selesai", "Tgl Temuan",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, r := range rows {
		statusStr := r.Status
		if r.Status == "RESOLVED" || r.Status == "CLOSED" {
			statusStr = "SELESAI"
		} else if r.Status == "IN_REVIEW" {
			statusStr = "DALAM REVIEW"
		} else if r.Status == "OPEN" {
			statusStr = "TERBUKA"
		}
		resDate := "-"
		if r.ReviewedAt != nil {
			resDate = r.ReviewedAt.Format("2006-01-02 15:04:05")
		}

		record := []string{
			r.DiscrepancyNumber,
			r.SimakNumber,
			r.WorkerName,
			r.ItemName,
			r.DiscrepancyType,
			fmt.Sprintf("%d", r.ExpectedQty),
			fmt.Sprintf("%d", r.ActualQty),
			statusStr,
			r.ResolutionNotes,
			resDate,
			r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}
