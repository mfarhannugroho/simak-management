package department

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

// Department represents plant internal department or workshop division.
type Department struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrganizationID *uuid.UUID `gorm:"type:uuid" json:"organization_id"`
	Name           string     `gorm:"type:varchar(100);not null" json:"name"`
	Code           string     `gorm:"type:varchar(20)" json:"code"`
	IsActive       bool       `gorm:"default:true;not null" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type DepartmentRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	IsActive *bool  `json:"is_active"`
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ListDepartments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	query := h.db.Model(&Department{})
	if search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var depts []Department
	offset := (page - 1) * limit
	if err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&depts).Error; err != nil {
		response.InternalServerError(c, "Gagal mengambil data departemen: "+err.Error())
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, "Data departemen berhasil dimuat", depts, response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID departemen tidak valid", err.Error())
		return
	}

	var dept Department
	if err := h.db.First(&dept, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Departemen tidak ditemukan")
			return
		}
		response.InternalServerError(c, "Database error: "+err.Error())
		return
	}

	response.Success(c, "Detail departemen berhasil dimuat", dept)
}

func (h *Handler) CreateDepartment(c *gin.Context) {
	var req DepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid", err.Error())
		return
	}

	dept := Department{
		Name:     req.Name,
		Code:     req.Code,
		IsActive: true,
	}
	if req.IsActive != nil {
		dept.IsActive = *req.IsActive
	}

	if err := h.db.Create(&dept).Error; err != nil {
		response.InternalServerError(c, "Gagal menyimpan departemen: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Departemen berhasil ditambahkan",
		Data:    dept,
	})
}

func (h *Handler) UpdateDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID departemen tidak valid", err.Error())
		return
	}

	var dept Department
	if err := h.db.First(&dept, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Departemen tidak ditemukan")
		return
	}

	var req DepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid", err.Error())
		return
	}

	dept.Name = req.Name
	dept.Code = req.Code
	if req.IsActive != nil {
		dept.IsActive = *req.IsActive
	}

	if err := h.db.Save(&dept).Error; err != nil {
		response.InternalServerError(c, "Gagal memperbarui departemen: "+err.Error())
		return
	}

	response.Success(c, "Departemen berhasil diperbarui", dept)
}

func (h *Handler) DeleteDepartment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID departemen tidak valid", err.Error())
		return
	}

	if err := h.db.Delete(&Department{}, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Gagal menghapus departemen: "+err.Error())
		return
	}

	response.Success(c, "Departemen berhasil dihapus", nil)
}
