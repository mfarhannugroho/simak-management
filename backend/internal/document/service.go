package document

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/simak"
)

type Service struct {
	db          *gorm.DB
	frontendURL string
}

func NewService(db *gorm.DB, frontendURL string) *Service {
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	return &Service{
		db:          db,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
}

// IssueSimak generates official document numbering, tamper-proof QR code, and transitions SIMAK to ISSUED
func (s *Service) IssueSimak(simakID uuid.UUID, officerID uuid.UUID, req IssueSimakRequest) (*SimakDocument, error) {
	var item simak.Simak
	err := s.db.
		Preload("WorkerProfile").
		Preload("WorkerProfile.Company").
		Preload("WorkerProfile.Department").
		Preload("Items").
		First(&item, "id = ?", simakID).Error
	if err != nil {
		return nil, errors.New("SIMAK tidak ditemukan")
	}

	// Must be APPROVED or already ISSUED (re-print)
	if item.Status != simak.StatusApproved && item.Status != simak.StatusIssued && item.Status != simak.StatusActive {
		return nil, fmt.Errorf("Hanya permohonan yang telah disetujui (APPROVED) yang dapat diterbitkan. Status saat ini: %s", item.Status)
	}

	now := time.Now()
	year := now.Format("2006")

	// 1. Concurrency-safe atomic generation of official SIMAK Number: SIMAK/INKA/YYYY/XXXXXX
	officialNumber := req.CustomNumber
	if officialNumber == "" {
		if item.SimakNumber != nil && strings.HasPrefix(*item.SimakNumber, fmt.Sprintf("SIMAK/INKA/%s/", year)) {
			officialNumber = *item.SimakNumber
		} else {
			prefix := fmt.Sprintf("SIMAK/INKA/%s/", year)
			var count int64
			s.db.Model(&simak.Simak{}).Where("simak_number LIKE ?", prefix+"%").Count(&count)
			officialNumber = fmt.Sprintf("%s%06d", prefix, count+1)
		}
	}

	// 2. Generate secure QR token if not already exists
	qrToken := ""
	if item.QRToken != nil && *item.QRToken != "" {
		qrToken = *item.QRToken
	} else {
		qrToken = GenerateQRToken()
	}

	// Public verification URL encoded into QR Code
	verificationURL := fmt.Sprintf("%s/verify/%s", s.frontendURL, qrToken)

	// Generate QR Code Base64 PNG image
	qrCodeData, err := GenerateQRCodeBase64(verificationURL)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat gambar QR Code: %w", err)
	}

	var doc SimakDocument

	// 3. Execute in DB Transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		item.Status = simak.StatusIssued
		item.SimakNumber = &officialNumber
		item.QRToken = &qrToken
		item.IssuedAt = &now
		item.IssuedBy = &officerID
		if req.Notes != "" {
			item.Notes = fmt.Sprintf("%s [Penerbitan: %s]", item.Notes, req.Notes)
		}

		if err := tx.Save(&item).Error; err != nil {
			return err
		}

		// Delete any existing document record for this SIMAK
		_ = tx.Where("simak_id = ?", simakID).Delete(&SimakDocument{}).Error

		fileName := fmt.Sprintf("SIMAK_%s.pdf", strings.ReplaceAll(strings.ReplaceAll(officialNumber, "/", "_"), " ", "_"))

		doc = SimakDocument{
			SimakID:        simakID,
			DocumentNumber: officialNumber,
			FileName:       fileName,
			FileURL:        fmt.Sprintf("/api/v1/simak/%s/document", simakID.String()),
			DocType:        "OFFICIAL_PDF",
			QRCodeData:     qrCodeData,
			IssuedBy:       &officerID,
			IssuedAt:       now,
		}

		return tx.Create(&doc).Error
	})

	if err != nil {
		return nil, err
	}

	return s.GetDocument(simakID)
}

// GetDocument retrieves the issued document metadata and QR code
func (s *Service) GetDocument(simakID uuid.UUID) (*SimakDocument, error) {
	var doc SimakDocument
	err := s.db.
		Preload("Simak").
		Preload("Simak.WorkerProfile").
		Preload("Simak.WorkerProfile.Company").
		Preload("Simak.WorkerProfile.Department").
		Preload("Simak.Items").
		Preload("Simak.Items.Category").
		Preload("Simak.Attachments").
		Preload("Issuer").
		First(&doc, "simak_id = ?", simakID).Error
	if err != nil {
		return nil, errors.New("Dokumen SIMAK resmi belum diterbitkan untuk pengajuan ini")
	}
	return &doc, nil
}

// VerifyQRToken evaluates public validity of a scanned QR token per PRD §16 & §17
func (s *Service) VerifyQRToken(qrToken string) (*PublicVerificationResponse, error) {
	var item simak.Simak
	err := s.db.
		Preload("WorkerProfile").
		Preload("WorkerProfile.Company").
		Preload("WorkerProfile.Department").
		Preload("Items").
		First(&item, "qr_token = ?", qrToken).Error

	if err != nil {
		return &PublicVerificationResponse{
			Valid:   false,
			Status:  "NOT_FOUND",
			Reason:  "Token verifikasi QR Code tidak ditemukan dalam sistem resmi PT INKA (Persero).",
		}, nil
	}

	simakNum := ""
	if item.SimakNumber != nil {
		simakNum = *item.SimakNumber
	}

	workerName := ""
	workerCode := ""
	companyName := "PT INKA (Persero)"
	deptName := ""
	position := ""

	if item.WorkerProfile != nil {
		workerName = item.WorkerProfile.FullName
		workerCode = item.WorkerProfile.WorkerID
		position = item.WorkerProfile.Position
		if item.WorkerProfile.Company != nil {
			companyName = item.WorkerProfile.Company.Name
		}
		if item.WorkerProfile.Department != nil {
			deptName = item.WorkerProfile.Department.Name
		}
	}

	issuedAtStr := ""
	if item.IssuedAt != nil {
		issuedAtStr = item.IssuedAt.Format("02 Jan 2006 15:04")
	}

	// Map items for verification checklist
	itemsList := make([]VerificationItemInfo, len(item.Items))
	for i, it := range item.Items {
		itemsList[i] = VerificationItemInfo{
			ItemName:     it.ItemName,
			Brand:        it.Brand,
			Model:        it.Model,
			SerialNumber: it.SerialNumber,
			Quantity:     it.Quantity,
			Unit:         it.Unit,
			Description:  it.Description,
		}
	}

	// 1. Cancellation check
	if item.Status == simak.StatusCancelled {
		return &PublicVerificationResponse{
			Valid:           false,
			Status:          "CANCELLED",
			Reason:          "Dokumen SIMAK ini telah DIBATALKAN secara resmi.",
			SimakNumber:     simakNum,
			ApplicantName:   workerName,
			Company:         companyName,
			DestinationArea: item.DestinationArea,
			StartDate:       item.StartDate.Format("02 Jan 2006"),
			EndDate:         item.EndDate.Format("02 Jan 2006"),
			TotalItems:      len(item.Items),
			Items:           itemsList,
		}, nil
	}

	// 2. Rejection check
	if item.Status == simak.StatusRejected {
		return &PublicVerificationResponse{
			Valid:           false,
			Status:          "INVALID",
			Reason:          "Permohonan SIMAK ini DITOLAK pada tahap pemeriksaan.",
			SimakNumber:     simakNum,
			ApplicantName:   workerName,
			Company:         companyName,
			DestinationArea: item.DestinationArea,
		}, nil
	}

	// 3. Expiration check
	today := time.Now().Truncate(24 * time.Hour)
	startDateTrunc := item.StartDate.Truncate(24 * time.Hour)
	endDateTrunc := item.EndDate.Truncate(24 * time.Hour)

	if today.After(endDateTrunc) {
		return &PublicVerificationResponse{
			Valid:           false,
			Status:          "EXPIRED",
			Reason:          fmt.Sprintf("Masa berlaku SIMAK telah kedaluwarsa pada tanggal %s.", item.EndDate.Format("02 January 2006")),
			SimakNumber:     simakNum,
			ApplicantName:   workerName,
			Company:         companyName,
			DestinationArea: item.DestinationArea,
			StartDate:       item.StartDate.Format("02 Jan 2006"),
			EndDate:         item.EndDate.Format("02 Jan 2006"),
			TotalItems:      len(item.Items),
			Items:           itemsList,
		}, nil
	}

	if today.Before(startDateTrunc) {
		return &PublicVerificationResponse{
			Valid:           false,
			Status:          "NOT_YET_VALID",
			Reason:          fmt.Sprintf("SIMAK ini baru akan mulai berlaku pada tanggal %s.", item.StartDate.Format("02 January 2006")),
			SimakNumber:     simakNum,
			ApplicantName:   workerName,
			Company:         companyName,
			DestinationArea: item.DestinationArea,
			StartDate:       item.StartDate.Format("02 Jan 2006"),
			EndDate:         item.EndDate.Format("02 Jan 2006"),
			TotalItems:      len(item.Items),
			Items:           itemsList,
		}, nil
	}

	// 4. Valid document!
	return &PublicVerificationResponse{
		Valid:           true,
		Status:          "VALID",
		SimakNumber:     simakNum,
		ApplicantName:   workerName,
		WorkerID:        workerCode,
		Company:         companyName,
		Department:      deptName,
		Position:        position,
		Purpose:         item.Purpose,
		Direction:       item.Direction,
		DestinationArea: item.DestinationArea,
		StartDate:       item.StartDate.Format("02 January 2006"),
		EndDate:         item.EndDate.Format("02 January 2006"),
		IssuedAt:        issuedAtStr,
		TotalItems:      len(item.Items),
		Items:           itemsList,
	}, nil
}

// GetIssuanceQueue retrieves SIMAKs for the officer dashboard
func (s *Service) GetIssuanceQueue(page int, limit int, status string, search string) ([]IssuanceQueueItem, int64, error) {
	query := s.db.Table("simak s").
		Select(`
			s.id,
			s.simak_number,
			s.purpose,
			s.direction,
			s.destination_area,
			s.start_date,
			s.end_date,
			s.status,
			s.qr_token,
			s.issued_at,
			s.created_at,
			w.full_name as worker_name,
			w.worker_id,
			COALESCE(c.name, 'PT INKA (Persero)') as company,
			COALESCE(d.name, '') as department,
			(SELECT COUNT(*) FROM simak_items si WHERE si.simak_id = s.id) as total_items
		`).
		Joins("JOIN workers w ON w.id = s.worker_id").
		Joins("LEFT JOIN companies c ON c.id = w.company_id").
		Joins("LEFT JOIN departments d ON d.id = w.department_id")

	if status != "" && status != "ALL" {
		query = query.Where("s.status = ?", status)
	} else {
		// By default show APPROVED, ISSUED, ACTIVE
		query = query.Where("s.status IN ('APPROVED', 'ISSUED', 'ACTIVE')")
	}

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("(LOWER(s.simak_number) LIKE ? OR LOWER(s.purpose) LIKE ? OR LOWER(w.full_name) LIKE ?)", searchTerm, searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var items []IssuanceQueueItem
	if err := query.Order("s.updated_at desc").Offset(offset).Limit(limit).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
