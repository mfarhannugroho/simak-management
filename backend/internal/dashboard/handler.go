package dashboard

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/internal/auth"
	"simak-backend/pkg/response"
)

type Handler struct {
	service *Service
	db      *gorm.DB
}

func NewHandler(service *Service, db *gorm.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
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

// GetSummary returns role-tailored KPI stats
// GET /api/v1/dashboard/summary
func (h *Handler) GetSummary(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	var user auth.User
	if err := h.db.Preload("Roles").Preload("Worker.Department").Preload("Worker.Company").First(&user, "id = ?", userID).Error; err != nil {
		response.NotFound(c, "Data pengguna tidak ditemukan")
		return
	}

	summary, err := h.service.GetSummary(&user)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat ringkasan dashboard: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// GetVolumeTrend returns 7-day movement trend
// GET /api/v1/dashboard/charts/volume-trend
func (h *Handler) GetVolumeTrend(c *gin.Context) {
	trend, err := h.service.GetVolumeTrend()
	if err != nil {
		response.InternalServerError(c, "Gagal memuat tren volume pergerakan: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    trend,
	})
}

// GetActivities returns recent system activities
// GET /api/v1/dashboard/activities
func (h *Handler) GetActivities(c *gin.Context) {
	period := c.DefaultQuery("period", "all")
	activities, err := h.service.GetActivities(period)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat aktivitas terkini: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
	})
}
