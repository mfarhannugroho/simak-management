package simak

import (
	"net/http"
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

// helper to get current worker id for the authenticated user
func (h *Handler) getWorkerIDForUser(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	var w struct {
		ID uuid.UUID
	}
	if err := h.db.Table("workers").Select("id").Where("user_id = ?", userIDVal).First(&w).Error; err != nil {
		return uuid.Nil, false
	}
	return w.ID, true
}

func isUserAdminOrOfficer(c *gin.Context) bool {
	rolesVal, exists := c.Get("roles")
	if !exists {
		return false
	}
	roles, ok := rolesVal.([]string)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == "Admin" || r == "Supervisor" || r == "SIMAK Officer" || r == "Approver" || r == "Gate Officer" {
			return true
		}
	}
	return false
}

// List handles GET /api/v1/simak
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := SimakFilterParams{
		Status:    c.Query("status"),
		Direction: c.Query("direction"),
		Search:    c.Query("search"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Page:      page,
		Limit:     limit,
	}

	isAdmin := isUserAdminOrOfficer(c)
	if !isAdmin {
		// Non-privileged users (Applicant) can only view their own SIMAK applications
		workerID, found := h.getWorkerIDForUser(c)
		if !found {
			response.NotFound(c, "Profil pekerja belum terdaftar untuk akun ini")
			return
		}
		params.WorkerID = workerID.String()
	} else if c.Query("worker_id") != "" {
		params.WorkerID = c.Query("worker_id")
	}

	list, total, err := h.service.List(params)
	if err != nil {
		response.InternalServerError(c, "Gagal memuat permohonan SIMAK: "+err.Error())
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	response.Success(c, "Daftar permohonan SIMAK berhasil dimuat", gin.H{
		"items": list,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

// GetByID handles GET /api/v1/simak/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", err.Error())
		return
	}

	item, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "Data SIMAK tidak ditemukan")
		return
	}

	// Security check for applicants
	if !isUserAdminOrOfficer(c) {
		workerID, found := h.getWorkerIDForUser(c)
		if !found || item.ApplicantID != workerID {
			response.Forbidden(c, "Anda tidak memiliki hak akses untuk melihat permohonan ini")
			return
		}
	}

	response.Success(c, "Detail SIMAK berhasil dimuat", item)
}

// Create handles POST /api/v1/simak
func (h *Handler) Create(c *gin.Context) {
	var req CreateSimakRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data permohonan tidak valid", err.Error())
		return
	}

	var workerID uuid.UUID
	if req.WorkerID != "" && isUserAdminOrOfficer(c) {
		parsed, err := uuid.Parse(req.WorkerID)
		if err == nil {
			workerID = parsed
		}
	}

	if workerID == uuid.Nil {
		var found bool
		workerID, found = h.getWorkerIDForUser(c)
		if !found {
			response.BadRequest(c, "Profil pekerja belum terdaftar. Harap lengkapi profil terlebih dahulu", nil)
			return
		}
	}

	simakObj, err := h.service.Create(workerID, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Permohonan SIMAK berhasil dibuat",
		Data:    simakObj,
	})
}

// Update handles PUT /api/v1/simak/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", err.Error())
		return
	}

	var req UpdateSimakRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Data permohonan tidak valid", err.Error())
		return
	}

	workerID, _ := h.getWorkerIDForUser(c)
	isAdmin := isUserAdminOrOfficer(c)

	updated, err := h.service.Update(id, workerID, isAdmin, req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Permohonan SIMAK berhasil diperbarui", updated)
}

// Delete handles DELETE /api/v1/simak/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", err.Error())
		return
	}

	workerID, _ := h.getWorkerIDForUser(c)
	isAdmin := isUserAdminOrOfficer(c)

	if err := h.service.Delete(id, workerID, isAdmin); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Draft permohonan SIMAK berhasil dihapus", nil)
}

// Submit handles POST /api/v1/simak/:id/submit
func (h *Handler) Submit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", err.Error())
		return
	}

	var req SubmitSimakRequest
	_ = c.ShouldBindJSON(&req)

	workerID, _ := h.getWorkerIDForUser(c)
	isAdmin := isUserAdminOrOfficer(c)

	submitted, err := h.service.Submit(id, workerID, isAdmin, req.Notes)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Permohonan SIMAK berhasil diajukan untuk ditinjau", submitted)
}

// Cancel handles POST /api/v1/simak/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", err.Error())
		return
	}

	var req CancelSimakRequest
	_ = c.ShouldBindJSON(&req)

	workerID, _ := h.getWorkerIDForUser(c)
	isAdmin := isUserAdminOrOfficer(c)

	cancelled, err := h.service.Cancel(id, workerID, isAdmin, req.Reason)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Success(c, "Permohonan SIMAK berhasil dibatalkan", cancelled)
}

// AddAttachment handles POST /api/v1/simak/:id/attachments
func (h *Handler) AddAttachment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID SIMAK tidak valid", err.Error())
		return
	}

	var input SimakAttachmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Data lampiran tidak valid", err.Error())
		return
	}

	att, err := h.service.AddAttachment(id, input)
	if err != nil {
		response.InternalServerError(c, "Gagal mengunggah lampiran: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Dokumen lampiran berhasil disimpan",
		Data:    att,
	})
}

// DeleteAttachment handles DELETE /api/v1/simak/attachments/:attachmentId
func (h *Handler) DeleteAttachment(c *gin.Context) {
	attID, err := uuid.Parse(c.Param("attachmentId"))
	if err != nil {
		response.BadRequest(c, "ID lampiran tidak valid", err.Error())
		return
	}

	if err := h.service.DeleteAttachment(attID); err != nil {
		response.InternalServerError(c, "Gagal menghapus lampiran: "+err.Error())
		return
	}

	response.Success(c, "Lampiran berhasil dihapus", nil)
}
