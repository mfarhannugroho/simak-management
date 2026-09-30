package company

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

// Company represents partner vendor, contractor, or internal INKA division.
type Company struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrganizationID *uuid.UUID `gorm:"type:uuid" json:"organization_id"`
	Name           string     `gorm:"type:varchar(150);not null" json:"name"`
	Type           string     `gorm:"type:varchar(20)" json:"type"` // INTERNAL, VENDOR, PT, CV
	Address        string     `gorm:"type:text" json:"address"`
	ContactPhone   string     `gorm:"type:varchar(30)" json:"contact_phone"`
	ContactEmail   string     `gorm:"type:varchar(100)" json:"contact_email"`
	IsActive       bool       `gorm:"default:true;not null" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Request DTO for creating/updating company
type CompanyRequest struct {
	Name         string `json:"name" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	IsActive     *bool  `json:"is_active"`
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// ListCompanies handles search, filtering, and pagination of companies.
func (h *Handler) ListCompanies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	companyType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	query := h.db.Model(&Company{})
	if search != "" {
		query = query.Where("name ILIKE ? OR address ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if companyType != "" {
		query = query.Where("type = ?", companyType)
	}

	var total int64
	query.Count(&total)

	var companies []Company
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&companies).Error; err != nil {
		response.InternalServerError(c, "Gagal mengambil data perusahaan: "+err.Error())
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, "Data perusahaan berhasil dimuat", companies, response.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	})
}

// GetCompany returns single company detail by ID.
func (h *Handler) GetCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID perusahaan tidak valid", err.Error())
		return
	}

	var comp Company
	if err := h.db.First(&comp, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Perusahaan tidak ditemukan")
			return
		}
		response.InternalServerError(c, "Database error: "+err.Error())
		return
	}

	response.Success(c, "Detail perusahaan berhasil dimuat", comp)
}

// CreateCompany creates a new company record.
func (h *Handler) CreateCompany(c *gin.Context) {
	var req CompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid", err.Error())
		return
	}

	comp := Company{
		Name:         req.Name,
		Type:         req.Type,
		Address:      req.Address,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		IsActive:     true,
	}
	if req.IsActive != nil {
		comp.IsActive = *req.IsActive
	}

	if err := h.db.Create(&comp).Error; err != nil {
		response.InternalServerError(c, "Gagal menyimpan perusahaan: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Perusahaan berhasil ditambahkan",
		Data:    comp,
	})
}

// UpdateCompany updates existing company.
func (h *Handler) UpdateCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID perusahaan tidak valid", err.Error())
		return
	}

	var comp Company
	if err := h.db.First(&comp, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Perusahaan tidak ditemukan")
		return
	}

	var req CompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data tidak valid", err.Error())
		return
	}

	comp.Name = req.Name
	comp.Type = req.Type
	comp.Address = req.Address
	comp.ContactPhone = req.ContactPhone
	comp.ContactEmail = req.ContactEmail
	if req.IsActive != nil {
		comp.IsActive = *req.IsActive
	}

	if err := h.db.Save(&comp).Error; err != nil {
		response.InternalServerError(c, "Gagal memperbarui perusahaan: "+err.Error())
		return
	}

	response.Success(c, "Perusahaan berhasil diperbarui", comp)
}

// DeleteCompany removes a company record.
func (h *Handler) DeleteCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID perusahaan tidak valid", err.Error())
		return
	}

	if err := h.db.Delete(&Company{}, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Gagal menghapus perusahaan: "+err.Error())
		return
	}

	response.Success(c, "Perusahaan berhasil dihapus", nil)
}
