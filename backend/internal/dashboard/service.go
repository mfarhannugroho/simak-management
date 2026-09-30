package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/auth"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetSummary returns role-tailored dashboard metrics
func (s *Service) GetSummary(user *auth.User) (*DashboardSummaryResponse, error) {
	primaryRole := "Applicant"
	if len(user.Roles) > 0 {
		primaryRole = user.Roles[0].Name
	}

	userName := user.Email
	if user.Worker != nil && user.Worker.FullName != "" {
		userName = user.Worker.FullName
	}

	resp := &DashboardSummaryResponse{
		Role:     primaryRole,
		UserName: userName,
		Metrics:  make([]RoleMetric, 0),
	}

	var workerID uuid.UUID
	if user.Worker != nil {
		workerID = user.Worker.ID
		resp.WorkerID = user.Worker.WorkerID
		resp.Position = user.Worker.Position
		if user.Worker.Department != nil {
			resp.Department = user.Worker.Department.Name
		}
		if user.Worker.Company != nil {
			resp.Company = user.Worker.Company.Name
		}
	}

	// Calculate overall SIMAK stats
	var totalSimak, activeSimak int64
	s.db.Table("simak").Count(&totalSimak)
	s.db.Table("simak").Where("status IN ?", []string{"ACTIVE", "ISSUED", "VERIFIED"}).Count(&activeSimak)
	resp.TotalSimak = totalSimak
	resp.ActiveSimak = activeSimak

	roleNorm := strings.ToUpper(strings.TrimSpace(primaryRole))

	switch {
	case strings.Contains(roleNorm, "APPLICANT") || strings.Contains(roleNorm, "PEKERJA"):
		var myActive, myPending, myCompleted int64
		wID := workerID
		if wID == uuid.Nil {
			wID = user.ID
		}
		s.db.Table("simak").Where("worker_id = ? AND status IN ?", wID, []string{"ACTIVE", "ISSUED", "VERIFIED"}).Count(&myActive)
		s.db.Table("simak").Where("worker_id = ? AND status IN ?", wID, []string{"SUBMITTED", "UNDER_REVIEW"}).Count(&myPending)
		s.db.Table("simak").Where("worker_id = ? AND status IN ?", wID, []string{"COMPLETED", "EXPIRED", "REJECTED"}).Count(&myCompleted)

		resp.Metrics = []RoleMetric{
			{
				Label:       "SIMAK Aktif",
				Value:       myActive,
				Description: "Surat izin peralatan aktif digunakan",
				Trend:       "+1 minggu ini",
				TrendType:   "positive",
				Highlight:   myActive > 0,
			},
			{
				Label:       "Menunggu Persetujuan",
				Value:       myPending,
				Description: "Sedang dalam proses verifikasi berjenjang",
				Trend:       "Dalam antrean",
				TrendType:   "neutral",
				Highlight:   false,
			},
			{
				Label:       "Riwayat Selesai",
				Value:       myCompleted,
				Description: "SIMAK yang telah selesai masa berlakunya",
				Trend:       "Total riwayat",
				TrendType:   "neutral",
				Highlight:   false,
			},
		}

	case strings.Contains(roleNorm, "APPROVER"):
		var pendingAction, approvedMonth, rejectedTotal int64
		s.db.Table("simak_approvals").
			Where("status = 'PENDING' AND (assigned_user_id = ? OR required_role = ?)", user.ID, primaryRole).
			Count(&pendingAction)

		startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
		s.db.Table("simak_approvals").
			Where("action_by_user_id = ? AND status = 'APPROVED' AND action_at >= ?", user.ID, startOfMonth).
			Count(&approvedMonth)

		s.db.Table("simak_approvals").
			Where("action_by_user_id = ? AND status = 'REJECTED'", user.ID).
			Count(&rejectedTotal)

		resp.Metrics = []RoleMetric{
			{
				Label:       "Menunggu Tindakan Anda",
				Value:       pendingAction,
				Description: "Pengajuan SIMAK butuh verifikasi Anda",
				Trend:       "Memerlukan tindakan",
				TrendType:   "negative",
				Highlight:   pendingAction > 0,
			},
			{
				Label:       "Disetujui Bulan Ini",
				Value:       approvedMonth,
				Description: "Pengajuan yang telah Anda setujui",
				Trend:       "Bulan berjalan",
				TrendType:   "positive",
				Highlight:   false,
			},
			{
				Label:       "Total Ditolak",
				Value:       rejectedTotal,
				Description: "Pengajuan yang tidak memenuhi syarat",
				Trend:       "Akumulatif",
				TrendType:   "neutral",
				Highlight:   false,
			},
		}

	case strings.Contains(roleNorm, "GATE") || strings.Contains(roleNorm, "SECURITY"):
		startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		var masukToday, keluarToday, discActive int64

		s.db.Table("inspections").
			Where("inspection_type = 'MASUK' AND inspected_at >= ?", startOfDay).
			Count(&masukToday)

		s.db.Table("inspections").
			Where("inspection_type = 'KELUAR' AND inspected_at >= ?", startOfDay).
			Count(&keluarToday)

		s.db.Table("discrepancies").
			Where("status IN ('OPEN', 'IN_REVIEW')").
			Count(&discActive)

		resp.Metrics = []RoleMetric{
			{
				Label:       "Masuk Hari Ini",
				Value:       masukToday,
				Description: "Verifikasi gerbang masuk personil & alat",
				Trend:       "Pergerakan hari ini",
				TrendType:   "positive",
				Highlight:   false,
			},
			{
				Label:       "Keluar Hari Ini",
				Value:       keluarToday,
				Description: "Verifikasi kepulangan dan pengembalian",
				Trend:       "Pergerakan hari ini",
				TrendType:   "neutral",
				Highlight:   false,
			},
			{
				Label:       "Temuan / Discrepancy",
				Value:       discActive,
				Description: "Ketidaksesuaian alat belum terselesaikan",
				Trend:       "Perlu perhatian",
				TrendType:   "negative",
				Highlight:   discActive > 0,
			},
		}

	case strings.Contains(roleNorm, "OFFICER"):
		var readyToIssue, totalAktif, inspectedToday int64
		s.db.Table("simak").Where("status = 'APPROVED'").Count(&readyToIssue)
		s.db.Table("simak").Where("status IN ?", []string{"ACTIVE", "ISSUED", "VERIFIED"}).Count(&totalAktif)
		startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		s.db.Table("inspections").Where("inspected_at >= ?", startOfDay).Count(&inspectedToday)

		resp.Metrics = []RoleMetric{
			{
				Label:       "Siap Diterbitkan",
				Value:       readyToIssue,
				Description: "Disetujui penuh & siap terbit QR",
				Trend:       "Siap diproses",
				TrendType:   "positive",
				Highlight:   readyToIssue > 0,
			},
			{
				Label:       "SIMAK Aktif Berjalan",
				Value:       totalAktif,
				Description: "Izin yang saat ini aktif di plant INKA",
				Trend:       "Aktif lapangan",
				TrendType:   "neutral",
				Highlight:   false,
			},
			{
				Label:       "Inspeksi Gerbang Hari Ini",
				Value:       inspectedToday,
				Description: "Fisik diverifikasi pos keamanan",
				Trend:       "Pemeriksaan hari ini",
				TrendType:   "positive",
				Highlight:   false,
			},
		}

	case strings.Contains(roleNorm, "SUPERVISOR"):
		var totalInsp, matchInsp, monthInsp, activeDisc int64
		s.db.Table("inspections").Count(&totalInsp)
		s.db.Table("inspections").Where("status = 'MATCH'").Count(&matchInsp)

		kepatuhan := int64(100)
		if totalInsp > 0 {
			kepatuhan = (matchInsp * 100) / totalInsp
		}

		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		s.db.Table("inspections").Where("inspected_at >= ?", thirtyDaysAgo).Count(&monthInsp)
		s.db.Table("discrepancies").Where("status IN ('OPEN', 'IN_REVIEW')").Count(&activeDisc)

		resp.Metrics = []RoleMetric{
			{
				Label:       "Tingkat Kepatuhan Fisik",
				Value:       kepatuhan,
				Description: fmt.Sprintf("%d dari %d inspeksi 100%% cocok", matchInsp, totalInsp),
				Trend:       fmt.Sprintf("%d%%", kepatuhan),
				TrendType:   "positive",
				Highlight:   kepatuhan < 90,
			},
			{
				Label:       "Pergerakan 30 Hari",
				Value:       monthInsp,
				Description: "Total lalu lintas peralatan di pos gerbang",
				Trend:       "Bulan berjalan",
				TrendType:   "neutral",
				Highlight:   false,
			},
			{
				Label:       "Temuan Belum Ditutup",
				Value:       activeDisc,
				Description: "Kasus ketidaksesuaian butuh penyelesaian",
				Trend:       "Audit alert",
				TrendType:   "negative",
				Highlight:   activeDisc > 0,
			},
		}

	default: // Superadmin / General
		var totalUsers, openDisc int64
		s.db.Table("users").Where("is_active = true").Count(&totalUsers)
		s.db.Table("discrepancies").Where("status IN ('OPEN', 'IN_REVIEW')").Count(&openDisc)

		resp.Metrics = []RoleMetric{
			{
				Label:       "Total SIMAK Sistem",
				Value:       totalSimak,
				Description: "Seluruh pengajuan SIMAK tercatat",
				Trend:       "Akumulatif",
				TrendType:   "positive",
				Highlight:   false,
			},
			{
				Label:       "Pengguna Aktif",
				Value:       totalUsers,
				Description: "Akun karyawan & vendor terdaftar aktif",
				Trend:       "Sistem operasional",
				TrendType:   "neutral",
				Highlight:   false,
			},
			{
				Label:       "Temuan Belum Selesai",
				Value:       openDisc,
				Description: "Discrepancy gerbang belum ditutup",
				Trend:       "Perlu peninjauan",
				TrendType:   "negative",
				Highlight:   openDisc > 0,
			},
		}
	}

	return resp, nil
}

// GetVolumeTrend returns 7-day movement trend data with peak day identification
func (s *Service) GetVolumeTrend() (*VolumeTrendResponse, error) {
	now := time.Now()
	// Indonesian day names mapping
	indonesianDays := map[time.Weekday]string{
		time.Monday:    "Sen",
		time.Tuesday:   "Sel",
		time.Wednesday: "Rab",
		time.Thursday:  "Kam",
		time.Friday:    "Jum",
		time.Saturday:  "Sab",
		time.Sunday:    "Min",
	}

	daily := make([]DailyVolumeItem, 7)
	var maxTotal int64 = -1
	var peakIndex int = 0
	var totalVolume int64 = 0

	// 6 days ago until today
	for i := 6; i >= 0; i-- {
		dayDate := now.AddDate(0, 0, -i)
		startOfDay := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), 0, 0, 0, 0, time.Local)
		endOfDay := startOfDay.Add(24 * time.Hour)

		var masuk, keluar int64
		s.db.Table("inspections").
			Where("inspection_type = 'MASUK' AND inspected_at >= ? AND inspected_at < ?", startOfDay, endOfDay).
			Count(&masuk)

		s.db.Table("inspections").
			Where("inspection_type = 'KELUAR' AND inspected_at >= ? AND inspected_at < ?", startOfDay, endOfDay).
			Count(&keluar)

		dayTotal := masuk + keluar
		totalVolume += dayTotal

		idx := 6 - i
		daily[idx] = DailyVolumeItem{
			Day:    indonesianDays[dayDate.Weekday()],
			Date:   dayDate.Format("2006-01-02"),
			Masuk:  masuk,
			Keluar: keluar,
			Total:  dayTotal,
			IsPeak: false,
		}

		if dayTotal > maxTotal {
			maxTotal = dayTotal
			peakIndex = idx
		}
	}

	if maxTotal > 0 {
		daily[peakIndex].IsPeak = true
	}

	return &VolumeTrendResponse{
		Period:      "7 Hari Terakhir",
		TotalVolume: totalVolume,
		PeakDay:     daily[peakIndex].Day,
		Daily:       daily,
	}, nil
}

// GetActivities returns recent system activities filtered by period
func (s *Service) GetActivities(period string) (*ActivitiesResponse, error) {
	var since time.Time
	now := time.Now()

	switch strings.ToLower(period) {
	case "today":
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	case "7d":
		since = now.AddDate(0, 0, -7)
	case "30d":
		since = now.AddDate(0, 0, -30)
	default:
		since = now.AddDate(0, 0, -90)
	}

	activities := make([]ActivityItem, 0)

	// 1. Recent Inspections
	type inspRow struct {
		ID             uuid.UUID
		InspectionType string
		Status         string
		GateName       string
		InspectedAt    time.Time
		SimakNumber    *string
	}
	var insps []inspRow
	s.db.Table("inspections").
		Select("inspections.id, inspections.inspection_type, inspections.status, inspections.gate_name, inspections.inspected_at, simak.simak_number").
		Joins("LEFT JOIN simak ON simak.id = inspections.simak_id").
		Where("inspections.inspected_at >= ?", since).
		Order("inspections.inspected_at desc").
		Limit(10).
		Scan(&insps)

	for _, in := range insps {
		num := "SIMAK"
		if in.SimakNumber != nil {
			num = *in.SimakNumber
		}
		variant := "inka"
		if in.Status == "MATCH" {
			variant = "success"
		} else if in.Status == "MISMATCH" {
			variant = "danger"
		}

		activities = append(activities, ActivityItem{
			ID:          in.ID.String(),
			Type:        "INSPECTION",
			Title:       fmt.Sprintf("Verifikasi Gerbang %s (%s)", in.InspectionType, num),
			Description: fmt.Sprintf("Diperiksa di %s • Status: %s", in.GateName, in.Status),
			Status:      in.Status,
			Variant:     variant,
			Timestamp:   in.InspectedAt,
			ActionURL:   "/gate/history",
		})
	}

	// 2. Recent Approvals
	type appRow struct {
		ID        uuid.UUID
		StepName  string
		Status    string
		ActionAt  *time.Time
		CreatedAt time.Time
		SimakNum  *string
	}
	var apps []appRow
	s.db.Table("simak_approvals").
		Select("simak_approvals.id, simak_approvals.step_name, simak_approvals.status, simak_approvals.action_at, simak_approvals.created_at, simak.simak_number as simak_num").
		Joins("LEFT JOIN simak ON simak.id = simak_approvals.simak_id").
		Where("simak_approvals.updated_at >= ?", since).
		Order("simak_approvals.updated_at desc").
		Limit(10).
		Scan(&apps)

	for _, ap := range apps {
		ts := ap.CreatedAt
		if ap.ActionAt != nil {
			ts = *ap.ActionAt
		}
		num := "SIMAK"
		if ap.SimakNum != nil {
			num = *ap.SimakNum
		}
		variant := "warning"
		if ap.Status == "APPROVED" {
			variant = "success"
		} else if ap.Status == "REJECTED" {
			variant = "danger"
		}

		activities = append(activities, ActivityItem{
			ID:          ap.ID.String(),
			Type:        "APPROVAL",
			Title:       fmt.Sprintf("Persetujuan %s: %s", ap.StepName, num),
			Description: fmt.Sprintf("Status tahap persetujuan: %s", ap.Status),
			Status:      ap.Status,
			Variant:     variant,
			Timestamp:   ts,
			ActionURL:   "/approvals",
		})
	}

	// 3. Recent Discrepancies
	type discRow struct {
		ID                uuid.UUID
		DiscrepancyNumber string
		ItemName          string
		DiscrepancyType   string
		Status            string
		CreatedAt         time.Time
	}
	var discs []discRow
	s.db.Table("discrepancies").
		Select("id, discrepancy_number, item_name, discrepancy_type, status, created_at").
		Where("created_at >= ?", since).
		Order("created_at desc").
		Limit(10).
		Scan(&discs)

	for _, d := range discs {
		status := "TERBUKA"
		variant := "danger"
		if d.Status == "RESOLVED" || d.Status == "CLOSED" {
			status = "SELESAI"
			variant = "success"
		}

		activities = append(activities, ActivityItem{
			ID:          d.ID.String(),
			Type:        "DISCREPANCY",
			Title:       fmt.Sprintf("Temuan Ketidaksesuaian (%s)", d.DiscrepancyNumber),
			Description: fmt.Sprintf("%s: %s", d.DiscrepancyType, d.ItemName),
			Status:      status,
			Variant:     variant,
			Timestamp:   d.CreatedAt,
			ActionURL:   "/discrepancies/" + d.ID.String(),
		})
	}

	// Sort activities descending by timestamp
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].Timestamp.Before(activities[j].Timestamp) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	if len(activities) > 15 {
		activities = activities[:15]
	}

	return &ActivitiesResponse{
		Period:     period,
		Activities: activities,
	}, nil
}
