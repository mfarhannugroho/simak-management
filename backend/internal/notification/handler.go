package notification

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"simak-backend/config"
	"simak-backend/pkg/response"
)

type Handler struct {
	service *Service
	cfg     *config.Config
}

func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

func parseUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("sesi otentikasi tidak ditemukan")
	}
	switch v := val.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, errors.New("format user_id tidak valid")
	}
}

// ListNotifications handles GET /api/v1/notifications
func (h *Handler) ListNotifications(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	unreadOnly, _ := strconv.ParseBool(c.DefaultQuery("unread_only", "false"))

	params := NotificationFilterParams{
		Page:       page,
		Limit:      limit,
		UnreadOnly: unreadOnly,
		Type:       c.Query("type"),
	}

	items, total, err := h.service.List(userID, params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat notifikasi: "+err.Error())
		return
	}

	pagination := response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	response.Paginated(c, "Daftar notifikasi berhasil dimuat", items, pagination)
}

// GetUnreadCount handles GET /api/v1/notifications/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	count, err := h.service.GetUnreadCount(userID)
	if err != nil {
		response.InternalServerError(c, "Gagal menghitung notifikasi baru: "+err.Error())
		return
	}

	response.Success(c, "Jumlah notifikasi belum dibaca berhasil dimuat", UnreadCountResponse{
		UnreadCount: count,
	})
}

// MarkAsRead handles PUT /api/v1/notifications/:id/read
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID notifikasi tidak valid", nil)
		return
	}

	notif, err := h.service.MarkAsRead(id, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Notifikasi telah ditandai dibaca", notif)
}

// MarkAllAsRead handles PUT /api/v1/notifications/read-all
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	if err := h.service.MarkAllAsRead(userID); err != nil {
		response.InternalServerError(c, "Gagal memperbarui status notifikasi: "+err.Error())
		return
	}

	response.Success(c, "Seluruh notifikasi telah ditandai dibaca", nil)
}

// HandleWebSocket handles GET /api/v1/ws (real-time notification stream)
func (h *Handler) HandleWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token otentikasi WebSocket tidak ditemukan"})
		return
	}

	// Validate JWT token
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritma signing token tidak valid")
		}
		return []byte(h.cfg.JWT.AccessSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token otentikasi WebSocket tidak sah"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Payload klaim token tidak valid"})
		return
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id tidak ditemukan dalam token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Format user_id tidak valid"})
		return
	}

	// Upgrade HTTP to WebSocket
	hub := h.service.GetHub()
	if hub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket Hub tidak aktif"})
		return
	}

	conn, err := hub.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("⚠️ WebSocket upgrade error: %v", err)
		return
	}

	hub.Register(userID, conn)

	// Keep alive loop reading client messages / pong
	go func() {
		defer hub.Unregister(userID, conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
