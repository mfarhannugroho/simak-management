package item

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"simak-backend/pkg/response"
)

// ItemCategory represents category classification for worker carried goods.
type ItemCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkerRegisteredItem represents personal work tools regularly carried by an applicant.
type WorkerRegisteredItem struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	WorkerID     uuid.UUID     `gorm:"type:uuid;not null" json:"worker_id"`
	CategoryID   *uuid.UUID    `gorm:"type:uuid" json:"category_id"`
	ItemName     string        `gorm:"type:varchar(150);not null" json:"item_name"`
	Brand        string        `gorm:"type:varchar(100)" json:"brand"`
	Model        string        `gorm:"type:varchar(100)" json:"model"`
	SerialNumber string        `gorm:"type:varchar(100)" json:"serial_number"`
	Quantity     int           `gorm:"default:1;not null" json:"quantity"`
	Unit         string        `gorm:"type:varchar(30);default:'Unit';not null" json:"unit"`
	Description  string        `gorm:"type:text" json:"description"`
	PhotoURL     string        `gorm:"type:text" json:"photo_url"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Category     *ItemCategory `gorm:"foreignKey:CategoryID" json:"category"`
}

type CategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type RegisteredItemRequest struct {
	WorkerID     string `json:"worker_id"`
	CategoryID   string `json:"category_id"`
	ItemName     string `json:"item_name" binding:"required"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
	Unit         string `json:"unit" binding:"required"`
	Description  string `json:"description"`
	PhotoURL     string `json:"photo_url"`
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// Item Categories Endpoints
// ─────────────────────────────────────────────────────────────────────────────

func (h *Handler) ListCategories(c *gin.Context) {
	search := c.Query("search")
	activeOnly := c.Query("active_only")

	query := h.db.Model(&ItemCategory{})
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if activeOnly == "true" {
		query = query.Where("is_active = true")
	}

	var categories []ItemCategory
	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		response.InternalServerError(c, "Gagal mengambil kategori barang: "+err.Error())
		return
	}

	response.Success(c, "Daftar kategori barang berhasil dimuat", categories)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data kategori tidak valid", err.Error())
		return
	}

	cat := ItemCategory{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}

	if err := h.db.Create(&cat).Error; err != nil {
		response.InternalServerError(c, "Gagal menambahkan kategori barang: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Kategori barang berhasil ditambahkan",
		Data:    cat,
	})
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID kategori tidak valid", err.Error())
		return
	}

	var cat ItemCategory
	if err := h.db.First(&cat, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Kategori tidak ditemukan")
		return
	}

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data kategori tidak valid", err.Error())
		return
	}

	cat.Name = req.Name
	cat.Description = req.Description
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}

	if err := h.db.Save(&cat).Error; err != nil {
		response.InternalServerError(c, "Gagal memperbarui kategori barang: "+err.Error())
		return
	}

	response.Success(c, "Kategori barang berhasil diperbarui", cat)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "ID kategori tidak valid", err.Error())
		return
	}

	if err := h.db.Delete(&ItemCategory{}, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Gagal menghapus kategori: "+err.Error())
		return
	}

	response.Success(c, "Kategori barang berhasil dihapus", nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// Worker Registered Items (Barang Bawaan Rutin Pekerja - PRD §7.1)
// ─────────────────────────────────────────────────────────────────────────────

func (h *Handler) ListWorkerItems(c *gin.Context) {
	workerIDParam := c.Param("id")

	// If empty or "my-items", check authenticated user's worker record
	var workerID uuid.UUID
	var err error

	if workerIDParam == "" || workerIDParam == "me" {
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "Sesi tidak ditemukan")
			return
		}
		var w struct {
			ID uuid.UUID
		}
		if err := h.db.Table("workers").Select("id").Where("user_id = ?", userID).First(&w).Error; err != nil {
			response.NotFound(c, "Profil pekerja belum terdaftar untuk akun ini")
			return
		}
		workerID = w.ID
	} else {
		workerID, err = uuid.Parse(workerIDParam)
		if err != nil {
			response.BadRequest(c, "ID pekerja tidak valid", err.Error())
			return
		}
	}

	var items []WorkerRegisteredItem
	if err := h.db.Preload("Category").Where("worker_id = ?", workerID).Order("created_at DESC").Find(&items).Error; err != nil {
		response.InternalServerError(c, "Gagal mengambil daftar barang rutin pekerja: "+err.Error())
		return
	}

	response.Success(c, "Daftar barang bawaan rutin berhasil dimuat", items)
}

func (h *Handler) CreateWorkerItem(c *gin.Context) {
	var req RegisteredItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data barang tidak valid", err.Error())
		return
	}

	var workerID uuid.UUID
	if req.WorkerID != "" {
		parsed, err := uuid.Parse(req.WorkerID)
		if err == nil {
			workerID = parsed
		}
	} else {
		// Default to authenticated user's worker
		userID, _ := c.Get("user_id")
		var w struct {
			ID uuid.UUID
		}
		if err := h.db.Table("workers").Select("id").Where("user_id = ?", userID).First(&w).Error; err == nil {
			workerID = w.ID
		}
	}

	if workerID == uuid.Nil {
		response.BadRequest(c, "Worker ID tidak valid", nil)
		return
	}

	item := WorkerRegisteredItem{
		WorkerID:     workerID,
		ItemName:     req.ItemName,
		Brand:        req.Brand,
		Model:        req.Model,
		SerialNumber: req.SerialNumber,
		Quantity:     req.Quantity,
		Unit:         req.Unit,
		Description:  req.Description,
		PhotoURL:     req.PhotoURL,
	}

	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			item.CategoryID = &catID
		}
	}

	if err := h.db.Create(&item).Error; err != nil {
		response.InternalServerError(c, "Gagal mendaftarkan barang rutin: "+err.Error())
		return
	}

	// Preload category for response
	_ = h.db.Preload("Category").First(&item, "id = ?", item.ID)

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Barang bawaan rutin berhasil didaftarkan",
		Data:    item,
	})
}

func (h *Handler) UpdateWorkerItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "ID barang tidak valid", err.Error())
		return
	}

	var item WorkerRegisteredItem
	if err := h.db.First(&item, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Data barang tidak ditemukan")
		return
	}

	var req RegisteredItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format data barang tidak valid", err.Error())
		return
	}

	item.ItemName = req.ItemName
	item.Brand = req.Brand
	item.Model = req.Model
	item.SerialNumber = req.SerialNumber
	item.Quantity = req.Quantity
	item.Unit = req.Unit
	item.Description = req.Description
	item.PhotoURL = req.PhotoURL

	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			item.CategoryID = &catID
		}
	}

	if err := h.db.Save(&item).Error; err != nil {
		response.InternalServerError(c, "Gagal memperbarui barang rutin: "+err.Error())
		return
	}

	_ = h.db.Preload("Category").First(&item, "id = ?", item.ID)
	response.Success(c, "Data barang bawaan berhasil diperbarui", item)
}

func (h *Handler) DeleteWorkerItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		response.BadRequest(c, "ID barang tidak valid", err.Error())
		return
	}

	if err := h.db.Delete(&WorkerRegisteredItem{}, "id = ?", id).Error; err != nil {
		response.InternalServerError(c, "Gagal menghapus barang: "+err.Error())
		return
	}

	response.Success(c, "Barang bawaan berhasil dihapus", nil)
}
