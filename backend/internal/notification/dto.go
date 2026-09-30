package notification

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotificationInput struct {
	UserID        uuid.UUID `json:"user_id"`
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	Type          string    `json:"type"`
	ReferenceID   string    `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	ActionURL     string    `json:"action_url"`
}

type NotificationFilterParams struct {
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=10"`
	UnreadOnly bool   `form:"unread_only"`
	Type       string `form:"type"`
}

type NotificationListItem struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Message       string     `json:"message"`
	Type          string     `json:"type"`
	ReferenceID   string     `json:"reference_id"`
	ReferenceType string     `json:"reference_type"`
	ActionURL     string     `json:"action_url"`
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
