package inspection

import (
	"errors"
	"strconv"

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

// RecordInspection handles POST /api/v1/verification/inspect
func (h *Handler) RecordInspection(c *gin.Context) {
	inspectorID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	var req CreateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data pemeriksaan fisik tidak valid: "+err.Error(), nil)
		return
	}

	ins, err := h.service.RecordInspection(inspectorID, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Created(c, "Hasil pemeriksaan fisik di pos gerbang berhasil disimpan", ins)
}

// ListInspections handles GET /api/v1/inspections
func (h *Handler) ListInspections(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := InspectionFilterParams{
		Page:           page,
		Limit:          limit,
		SimakID:        c.Query("simak_id"),
		GateName:       c.Query("gate_name"),
		InspectionType: c.Query("inspection_type"),
		Status:         c.Query("status"),
		Search:         c.Query("search"),
		Date:           c.Query("date"),
	}

	items, total, err := h.service.ListInspections(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat riwayat pemeriksaan pos gerbang: "+err.Error())
		return
	}

	pagination := response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	response.Paginated(c, "Riwayat pemeriksaan pos gerbang berhasil dimuat", items, pagination)
}

// GetInspection handles GET /api/v1/inspections/:id
func (h *Handler) GetInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID pemeriksaan tidak valid", nil)
		return
	}

	ins, err := h.service.GetInspectionByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Detail pemeriksaan pos gerbang berhasil dimuat", ins)
}

// GetMetrics handles GET /api/v1/gate/metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.service.GetGateDashboardMetrics()
	if err != nil {
		response.InternalServerError(c, "Gagal memuat metrik pos gerbang: "+err.Error())
		return
	}

	response.Success(c, "Metrik operasional pos gerbang berhasil dimuat", metrics)
}
