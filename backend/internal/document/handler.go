package document

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

// IssueSimak handles POST /api/v1/simak/:id/issue
func (h *Handler) IssueSimak(c *gin.Context) {
	officerID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	var req IssueSimakRequest
	_ = c.ShouldBindJSON(&req)

	doc, err := h.service.IssueSimak(simakID, officerID, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Dokumen resmi SIMAK dan QR Code verifikasi berhasil diterbitkan", doc)
}

// GetDocument handles GET /api/v1/simak/:id/document
func (h *Handler) GetDocument(c *gin.Context) {
	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	doc, err := h.service.GetDocument(simakID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Dokumen SIMAK resmi berhasil dimuat", doc)
}

// PublicVerifyToken handles public GET /api/v1/verification/:qrToken (No auth required)
func (h *Handler) PublicVerifyToken(c *gin.Context) {
	token := c.Param("qrToken")
	if token == "" {
		response.BadRequest(c, "Token verifikasi QR wajib disertakan", nil)
		return
	}

	res, err := h.service.VerifyQRToken(token)
	if err != nil {
		response.InternalServerError(c, "Gagal memverifikasi token QR: "+err.Error())
		return
	}

	response.Success(c, "Hasil verifikasi keaslian SIMAK", res)
}

// GetIssuanceQueue handles GET /api/v1/officer/issuance
func (h *Handler) GetIssuanceQueue(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.DefaultQuery("status", "APPROVED")
	search := c.Query("search")

	items, total, err := h.service.GetIssuanceQueue(page, limit, status, search)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat daftar permohonan siap terbit: "+err.Error())
		return
	}

	pagination := response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	response.Paginated(c, "Daftar permohonan penerbitan SIMAK berhasil dimuat", items, pagination)
}
