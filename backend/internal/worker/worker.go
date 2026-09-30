package worker

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/pkg/response"
)

type Worker struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID       uuid.UUID   `gorm:"type:uuid;unique;not null" json:"user_id"`
	CompanyID    *uuid.UUID  `gorm:"type:uuid" json:"company_id"`
	DepartmentID *uuid.UUID  `gorm:"type:uuid" json:"department_id"`
	WorkerID     string      `gorm:"type:varchar(50);unique" json:"worker_id"`
	FullName     string      `gorm:"type:varchar(150);not null" json:"full_name"`
	Position     string      `gorm:"type:varchar(100)" json:"position"`
	Phone        string      `gorm:"type:varchar(30)" json:"phone"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Company      *Company    `gorm:"foreignKey:CompanyID" json:"company"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department"`
}

type Company struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

type Department struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type WorkerRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	CompanyID    string `json:"company_id"`
	DepartmentID string `json:"department_id"`
	WorkerID     string `json:"worker_id" binding:"required"`
	FullName     string `json:"full_name" binding:"required"`
	Position     string `json:"position"`
	Phone        string `json:"phone"`
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ListWorkers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	companyID := c.Query("company_id")
	departmentID := c.Query("department_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	query := h.db.Model(&Worker{})
	if search != "" {
		query = query.Where("full_name ILIKE ? OR worker_id ILIKE ? OR phone ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}

	var total int64
	query.Count(&total)

	var workers []Worker
	offset := (page - 1) * limit
	if err := query.Preload("Company").Preload("Department").Order("created_at DESC").Offset(offset).Limit(limit).Find(&workers).Error; err != nil {
		response.InternalServerError(c, "Gagal mengambil data pekerja: "+err.Error())
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, "Data pekerja berhasil dimuat", workers, response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetWorker(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID pekerja tidak valid", err.Error())
		return
	}

	var w Worker
	if err := h.db.Preload("Company").Preload("Department").First(&w, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Data pekerja tidak ditemukan")
			return
		}
		response.InternalServerError(c, "Database error: "+err.Error())
		return
	}

	response.Success(c, "Detail pekerja berhasil dimuat", w)
}

func (h *Handler) CreateWorker(c *gin.Context) {
	var req WorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "User ID tidak valid", err.Error())
		return
	}

	w := Worker{
		UserID:   uid,
		WorkerID: req.WorkerID,
		FullName: req.FullName,
		Position: req.Position,
		Phone:    req.Phone,
	}

	if req.CompanyID != "" {
		cid, err := uuid.Parse(req.CompanyID)
		if err == nil {
			w.CompanyID = &cid
		}
	}
	if req.DepartmentID != "" {
		did, err := uuid.Parse(req.DepartmentID)
		if err == nil {
			w.DepartmentID = &did
		}
	}

	if err := h.db.Create(&w).Error; err != nil {
		response.InternalServerError(c, "Gagal menyimpan data pekerja: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Profil pekerja berhasil dibuat",
		Data:    w,
	})
}

func (h *Handler) UpdateWorker(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID pekerja tidak valid", err.Error())
		return
	}

	var w Worker
	if err := h.db.First(&w, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Data pekerja tidak ditemukan")
		return
	}

	var req WorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid", err.Error())
		return
	}

	w.WorkerID = req.WorkerID
	w.FullName = req.FullName
	w.Position = req.Position
	w.Phone = req.Phone

	if req.CompanyID != "" {
		cid, err := uuid.Parse(req.CompanyID)
		if err == nil {
			w.CompanyID = &cid
		}
	}
	if req.DepartmentID != "" {
		did, err := uuid.Parse(req.DepartmentID)
		if err == nil {
			w.DepartmentID = &did
		}
	}

	if err := h.db.Save(&w).Error; err != nil {
		response.InternalServerError(c, "Gagal memperbarui profil pekerja: "+err.Error())
		return
	}

	response.Success(c, "Profil pekerja berhasil diperbarui", w)
}

func (h *Handler) DeleteWorker(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID pekerja tidak valid", err.Error())
		return
	}

	if err := h.db.Delete(&Worker{}, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Gagal menghapus data pekerja: "+err.Error())
		return
	}

	response.Success(c, "Data pekerja berhasil dihapus", nil)
}
