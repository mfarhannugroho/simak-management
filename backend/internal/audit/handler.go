package audit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/pkg/response"
)

type Handler struct {
	service *Service
	db      *gorm.DB
}

func NewHandler(service *Service, db *gorm.DB) *Handler {
	return &Handler{service: service, db: db}
}

// ListAuditLogs handles GET /api/v1/audit-logs
func (h *Handler) ListAuditLogs(c *gin.Context) {
	var params AuditFilterParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Parameter filter audit tidak valid: "+err.Error(), nil)
		return
	}

	logs, total, err := h.service.ListAuditLogs(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat jejak audit: "+err.Error())
		return
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ExportCSV handles GET /api/v1/audit-logs/export
func (h *Handler) ExportCSV(c *gin.Context) {
	var params AuditFilterParams
	_ = c.ShouldBindQuery(&params)

	data, err := h.service.ExportAuditCSV(params)
	if err != nil {
		response.InternalServerError(c, "Gagal mengekspor riwayat audit: "+err.Error())
		return
	}

	filename := fmt.Sprintf("audit_trail_simak_inka_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// GetByID handles GET /api/v1/audit-logs/:id
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID audit log tidak valid", nil)
		return
	}

	var entry AuditLog
	if err := h.db.First(&entry, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Catatan audit log tidak ditemukan")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entry,
	})
}
