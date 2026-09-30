package discrepancy

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

// ListDiscrepancies handles GET /api/v1/discrepancies
func (h *Handler) ListDiscrepancies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := DiscrepancyFilterParams{
		Page:            page,
		Limit:           limit,
		Status:          c.Query("status"),
		DiscrepancyType: c.Query("discrepancy_type"),
		SimakID:         c.Query("simak_id"),
		Search:          c.Query("search"),
		Date:            c.Query("date"),
	}

	items, total, err := h.service.ListDiscrepancies(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar temuan ketidaksesuaian: "+err.Error())
		return
	}

	pagination := response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	response.Paginated(c, "Daftar ketidaksesuaian berhasil dimuat", items, pagination)
}

// GetDiscrepancy handles GET /api/v1/discrepancies/:id
func (h *Handler) GetDiscrepancy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID ketidaksesuaian tidak valid", nil)
		return
	}

	disc, err := h.service.GetDiscrepancyByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Detail temuan ketidaksesuaian berhasil dimuat", disc)
}

// GetMetrics handles GET /api/v1/discrepancies/metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	metrics, err := h.service.GetMetrics()
	if err != nil {
		response.InternalServerError(c, "Gagal memuat statistik ketidaksesuaian: "+err.Error())
		return
	}

	response.Success(c, "Statistik ketidaksesuaian berhasil dimuat", metrics)
}

// CreateDiscrepancy handles POST /api/v1/discrepancies
func (h *Handler) CreateDiscrepancy(c *gin.Context) {
	reporterID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	var req CreateDiscrepancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data laporan ketidaksesuaian tidak valid: "+err.Error(), nil)
		return
	}

	disc, err := h.service.CreateDiscrepancy(reporterID, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Created(c, "Laporan ketidaksesuaian barang berhasil dicatat", disc)
}

// StartReview handles PUT /api/v1/discrepancies/:id/review
func (h *Handler) StartReview(c *gin.Context) {
	supervisorID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID ketidaksesuaian tidak valid", nil)
		return
	}

	disc, err := h.service.StartReview(id, supervisorID)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Status ketidaksesuaian berhasil diubah menjadi Dalam Investigasi (IN_REVIEW)", disc)
}

// ResolveDiscrepancy handles PUT /api/v1/discrepancies/:id/resolve
func (h *Handler) ResolveDiscrepancy(c *gin.Context) {
	supervisorID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID ketidaksesuaian tidak valid", nil)
		return
	}

	var req ResolveDiscrepancyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data penyelesaian ketidaksesuaian tidak valid: "+err.Error(), nil)
		return
	}

	disc, err := h.service.ResolveDiscrepancy(id, supervisorID, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Kasus ketidaksesuaian barang berhasil diselesaikan", disc)
}
