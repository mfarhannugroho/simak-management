package notification

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/simak"
)

type Service struct {
	db  *gorm.DB
	hub *Hub
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	return &Service{
		db:  db,
		hub: hub,
	}
}

func (s *Service) GetHub() *Hub {
	return s.hub
}

// Send saves an in-app notification and immediately broadcasts it to online WebSocket clients
func (s *Service) Send(input CreateNotificationInput) (*Notification, error) {
	if input.UserID == uuid.Nil {
		return nil, errors.New("user_id penerima notifikasi tidak valid")
	}

	notif := Notification{
		UserID:        input.UserID,
		Title:         input.Title,
		Message:       input.Message,
		Type:          input.Type,
		ReferenceID:   input.ReferenceID,
		ReferenceType: input.ReferenceType,
		ActionURL:     input.ActionURL,
		IsRead:        false,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(&notif).Error; err != nil {
		return nil, err
	}

	// Real-time broadcast
	if s.hub != nil {
		s.hub.BroadcastToUser(input.UserID, notif)
	}

	return &notif, nil
}

// SendToRole dispatches notifications to all active users with a specified role
func (s *Service) SendToRole(roleName string, input CreateNotificationInput) error {
	var userIDs []uuid.UUID
	err := s.db.Table("users u").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("r.name = ? AND u.is_active = true", roleName).
		Pluck("u.id", &userIDs).Error
	if err != nil {
		return err
	}

	for _, uid := range userIDs {
		userInp := input
		userInp.UserID = uid
		_, _ = s.Send(userInp)
	}

	return nil
}

// List returns paginated notifications for the specified user
func (s *Service) List(userID uuid.UUID, p NotificationFilterParams) ([]NotificationListItem, int64, error) {
	query := s.db.Model(&Notification{}).Where("user_id = ?", userID)

	if p.UnreadOnly {
		query = query.Where("is_read = false")
	}

	if p.Type != "" && p.Type != "ALL" {
		query = query.Where("type = ?", p.Type)
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

	var items []NotificationListItem
	err := query.Order("created_at desc").
		Offset(offset).
		Limit(p.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetUnreadCount calculates the count of unread notifications for a user
func (s *Service) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a specific notification as read
func (s *Service) MarkAsRead(id uuid.UUID, userID uuid.UUID) (*Notification, error) {
	var notif Notification
	if err := s.db.First(&notif, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, errors.New("Notifikasi tidak ditemukan")
	}

	now := time.Now()
	notif.IsRead = true
	notif.ReadAt = &now

	if err := s.db.Save(&notif).Error; err != nil {
		return nil, err
	}

	return &notif, nil
}

// MarkAllAsRead marks all unread notifications of a user as read
func (s *Service) MarkAllAsRead(userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// CheckExpiringAndExpiredSimaks runs scheduled checks for validity lifecycle
func (s *Service) CheckExpiringAndExpiredSimaks() error {
	today := time.Now().Truncate(24 * time.Hour)
	threeDaysLater := today.Add(3 * 24 * time.Hour)

	// 1. Check expired SIMAKs
	var expiredList []simak.Simak
	err := s.db.
		Preload("WorkerProfile").
		Where("end_date < ? AND status IN ('ACTIVE', 'ISSUED', 'VERIFIED')", today).
		Find(&expiredList).Error
	if err == nil {
		for _, item := range expiredList {
			item.Status = simak.StatusExpired
			s.db.Save(&item)

			if item.WorkerProfile != nil {
				_, _ = s.Send(CreateNotificationInput{
					UserID:        item.WorkerProfile.UserID,
					Title:         "Masa Berlaku SIMAK Telah Berakhir",
					Message:       fmt.Sprintf("Izin membawa alat kerja nomor %s telah kedaluwarsa per tanggal %s.", item.SimakNumber, item.EndDate.Format("02/01/2006")),
					Type:          TypeSimakExpired,
					ReferenceID:   item.ID.String(),
					ReferenceType: "simak",
					ActionURL:     fmt.Sprintf("/simak/%s", item.ID),
				})
			}
		}
	}

	// 2. Check expiring SIMAKs (within 3 days)
	var expiringList []simak.Simak
	err = s.db.
		Preload("WorkerProfile").
		Where("end_date >= ? AND end_date <= ? AND status IN ('ACTIVE', 'ISSUED', 'VERIFIED')", today, threeDaysLater).
		Find(&expiringList).Error
	if err == nil {
		for _, item := range expiringList {
			if item.WorkerProfile != nil {
				// Avoid duplicate notification on same day
				var existing int64
				s.db.Model(&Notification{}).
					Where("user_id = ? AND reference_id = ? AND type = ? AND DATE(created_at) = ?",
						item.WorkerProfile.UserID, item.ID.String(), TypeSimakExpiring, today.Format("2006-01-02")).
					Count(&existing)

				if existing == 0 {
					_, _ = s.Send(CreateNotificationInput{
						UserID:        item.WorkerProfile.UserID,
						Title:         "Masa Berlaku SIMAK Segera Berakhir",
						Message:       fmt.Sprintf("Izin membawa alat nomor %s akan berakhir pada %s. Mohon persiapkan rekonsiliasi pengeluaran alat.", item.SimakNumber, item.EndDate.Format("02/01/2006")),
						Type:          TypeSimakExpiring,
						ReferenceID:   item.ID.String(),
						ReferenceType: "simak",
						ActionURL:     fmt.Sprintf("/simak/%s", item.ID),
					})
				}
			}
		}
	}

	return nil
}

// StartCronScheduler starts a lightweight background ticker in a goroutine
func (s *Service) StartCronScheduler(interval time.Duration) {
	go func() {
		log.Printf("⏰ Notification Cron Scheduler started (interval: %v)", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run once on startup
		if err := s.CheckExpiringAndExpiredSimaks(); err != nil {
			log.Printf("⚠️ Cron scheduler run error: %v", err)
		}

		for range ticker.C {
			if err := s.CheckExpiringAndExpiredSimaks(); err != nil {
				log.Printf("⚠️ Cron scheduler periodic error: %v", err)
			}
		}
	}()
}
