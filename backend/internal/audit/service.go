package audit

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Record inserts an immutable audit log entry
func (s *Service) Record(input CreateAuditInput) error {
	detailsStr := "{}"
	if input.Details != nil {
		if bytesVal, err := json.Marshal(input.Details); err == nil {
			detailsStr = string(bytesVal)
		}
	}

	logEntry := AuditLog{
		UserID:     input.UserID,
		UserEmail:  input.UserEmail,
		UserName:   input.UserName,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Details:    detailsStr,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		CreatedAt:  time.Now(),
	}

	return s.db.Create(&logEntry).Error
}

// ListAuditLogs returns paginated audit records with search and filters
func (s *Service) ListAuditLogs(params AuditFilterParams) ([]AuditLog, int64, error) {
	query := s.db.Model(&AuditLog{})

	if params.Action != "" {
		query = query.Where("action = ?", params.Action)
	}
	if params.EntityType != "" {
		query = query.Where("entity_type = ?", params.EntityType)
	}
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}
	if params.Search != "" {
		pattern := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where("LOWER(action) LIKE ? OR LOWER(entity_id) LIKE ? OR LOWER(user_name) LIKE ? OR LOWER(user_email) LIKE ? OR LOWER(ip_address) LIKE ?", pattern, pattern, pattern, pattern, pattern)
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

	var logs []AuditLog
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ExportAuditCSV exports audit logs into UTF-8 BOM CSV bytes
func (s *Service) ExportAuditCSV(params AuditFilterParams) ([]byte, error) {
	params.Page = 1
	params.PageSize = 10000
	logs, _, err := s.ListAuditLogs(params)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM for Microsoft Excel

	writer := csv.NewWriter(&buf)
	header := []string{
		"ID Audit", "Waktu (WIB)", "Aktor / Pengguna", "Email", "Aksi Sensitif",
		"Tipe Entitas", "ID Dokumen / Kasus", "Alamat IP", "User Agent", "Detail Mutasi",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, l := range logs {
		record := []string{
			l.ID.String(),
			l.CreatedAt.Format("2006-01-02 15:04:05"),
			l.UserName,
			l.UserEmail,
			l.Action,
			l.EntityType,
			l.EntityID,
			l.IPAddress,
			l.UserAgent,
			l.Details,
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), nil
}
