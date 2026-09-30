package report

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simak-backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetSimakReport handles GET /api/v1/reports/simak
func (h *Handler) GetSimakReport(c *gin.Context) {
	var params ReportFilterParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Parameter filter tidak valid: "+err.Error(), nil)
		return
	}

	rows, total, err := h.service.GetSimakReport(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat laporan SIMAK: "+err.Error())
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
		"data":    rows,
		"meta": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetVerificationReport handles GET /api/v1/reports/verification
func (h *Handler) GetVerificationReport(c *gin.Context) {
	var params ReportFilterParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Parameter filter tidak valid: "+err.Error(), nil)
		return
	}

	rows, total, err := h.service.GetVerificationReport(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat laporan verifikasi: "+err.Error())
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
		"data":    rows,
		"meta": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetDiscrepancyReport handles GET /api/v1/reports/discrepancies
func (h *Handler) GetDiscrepancyReport(c *gin.Context) {
	var params ReportFilterParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Parameter filter tidak valid: "+err.Error(), nil)
		return
	}

	rows, total, err := h.service.GetDiscrepancyReport(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat laporan ketidaksesuaian: "+err.Error())
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
		"data":    rows,
		"meta": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ExportCSV handles GET /api/v1/reports/export
func (h *Handler) ExportCSV(c *gin.Context) {
	reportType := c.DefaultQuery("type", "simak")
	var params ReportFilterParams
	_ = c.ShouldBindQuery(&params)

	var data []byte
	var filename string
	var err error

	timestamp := time.Now().Format("20060102_150405")

	switch reportType {
	case "verification", "gate":
		data, err = h.service.ExportVerificationCSV(params)
		filename = fmt.Sprintf("laporan_verifikasi_gerbang_%s.csv", timestamp)
	case "discrepancy", "temuan":
		data, err = h.service.ExportDiscrepancyCSV(params)
		filename = fmt.Sprintf("laporan_temuan_discrepancy_%s.csv", timestamp)
	default:
		data, err = h.service.ExportSimakCSV(params)
		filename = fmt.Sprintf("laporan_simak_%s.csv", timestamp)
	}

	if err != nil {
		response.InternalServerError(c, "Gagal menghasilkan berkas ekspor: "+err.Error())
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}
