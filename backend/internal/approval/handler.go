package approval

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

func parseUserRoles(c *gin.Context) []string {
	val, exists := c.Get("roles")
	if !exists {
		return []string{}
	}
	if roles, ok := val.([]string); ok {
		return roles
	}
	return []string{}
}

// GetPendingQueue handles GET /api/v1/approvals/pending
func (h *Handler) GetPendingQueue(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roles := parseUserRoles(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.DefaultQuery("status", "PENDING")

	query := PendingApprovalQuery{
		Page:   page,
		Limit:  limit,
		Search: search,
		Status: status,
	}

	items, total, err := h.service.GetPendingApprovals(userID, roles, query)
	if err != nil {
		response.InternalServerError(c, "Gagal mengambil daftar antrean persetujuan: "+err.Error())
		return
	}

	pagination := response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}

	response.Paginated(c, "Daftar antrean persetujuan berhasil dimuat", items, pagination)
}

// GetApprovalDetail handles GET /api/v1/approvals/detail/:id
func (h *Handler) GetApprovalDetail(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roles := parseUserRoles(c)

	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	detail, err := h.service.GetApprovalDetail(simakID, userID, roles)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, "Detail permohonan dan alur persetujuan berhasil dimuat", detail)
}

// ProcessDecision handles POST /api/v1/approvals/:id/decision
func (h *Handler) ProcessDecision(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roles := parseUserRoles(c)

	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	var req ApprovalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data keputusan tidak valid: "+err.Error(), nil)
		return
	}

	updatedSimak, err := h.service.ProcessDecision(simakID, userID, roles, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Keputusan persetujuan berhasil disimpan", updatedSimak)
}

// Approve handles POST /api/v1/simak/:id/approve
func (h *Handler) Approve(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roles := parseUserRoles(c)

	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	var body struct {
		Comments string `json:"comments"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Comments == "" {
		body.Comments = "Disetujui sesuai ketentuan yang berlaku"
	}

	req := ApprovalDecisionRequest{
		Decision: "APPROVE",
		Comments: body.Comments,
	}

	updated, err := h.service.ProcessDecision(simakID, userID, roles, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Permohonan SIMAK berhasil disetujui", updated)
}

// Reject handles POST /api/v1/simak/:id/reject
func (h *Handler) Reject(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roles := parseUserRoles(c)

	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	var body struct {
		Comments string `json:"comments" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Alasan penolakan wajib diisi", nil)
		return
	}

	req := ApprovalDecisionRequest{
		Decision: "REJECT",
		Comments: body.Comments,
	}

	updated, err := h.service.ProcessDecision(simakID, userID, roles, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Permohonan SIMAK telah ditolak", updated)
}

// RequestRevision handles POST /api/v1/simak/:id/revision
func (h *Handler) RequestRevision(c *gin.Context) {
	userID, err := parseUserID(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	roles := parseUserRoles(c)

	simakID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", nil)
		return
	}

	var body struct {
		Comments string `json:"comments" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Catatan revisi wajib diisi", nil)
		return
	}

	req := ApprovalDecisionRequest{
		Decision: "REQUEST_REVISION",
		Comments: body.Comments,
	}

	updated, err := h.service.ProcessDecision(simakID, userID, roles, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Permintaan revisi SIMAK berhasil dikirim ke pemohon", updated)
}

// ─────────────────────────────────────────────────────────────────────────────
// Admin Approval Matrix Handlers
// ─────────────────────────────────────────────────────────────────────────────

func (h *Handler) ListMatrices(c *gin.Context) {
	matrices, err := h.service.ListMatrices()
	if err != nil {
		response.InternalServerError(c, "Gagal memuat matriks persetujuan: "+err.Error())
		return
	}
	response.Success(c, "Daftar matriks persetujuan berhasil dimuat", matrices)
}

func (h *Handler) GetMatrix(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID matriks tidak valid", nil)
		return
	}
	matrix, err := h.service.GetMatrixByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, "Detail matriks persetujuan berhasil dimuat", matrix)
}

func (h *Handler) CreateMatrix(c *gin.Context) {
	var req CreateMatrixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data matriks persetujuan tidak valid: "+err.Error(), nil)
		return
	}

	matrix, err := h.service.CreateMatrix(req)
	if err != nil {
		response.BadRequest(c, "Gagal membuat matriks persetujuan: "+err.Error(), nil)
		return
	}

	response.Created(c, "Matriks persetujuan berhasil dibuat", matrix)
}

func (h *Handler) UpdateMatrix(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID matriks tidak valid", nil)
		return
	}

	var req UpdateMatrixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data matriks persetujuan tidak valid: "+err.Error(), nil)
		return
	}

	matrix, err := h.service.UpdateMatrix(id, req)
	if err != nil {
		response.BadRequest(c, "Gagal memperbarui matriks persetujuan: "+err.Error(), nil)
		return
	}

	response.Success(c, "Matriks persetujuan berhasil diperbarui", matrix)
}

func (h *Handler) DeleteMatrix(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID matriks tidak valid", nil)
		return
	}

	if err := h.service.DeleteMatrix(id); err != nil {
		response.BadRequest(c, "Gagal menghapus matriks persetujuan: "+err.Error(), nil)
		return
	}

	response.Success(c, "Matriks persetujuan berhasil dihapus", nil)
}
