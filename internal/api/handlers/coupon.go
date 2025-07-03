package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/toanthaycong_golang/internal/api/dto"
)

type CouponHandler struct {
	db *sql.DB
}

func NewCouponHandler(db *sql.DB) *CouponHandler {
	return &CouponHandler{db: db}
}

// GetCoupons godoc
// @Summary Lấy danh sách mã giảm giá
// @Description Lấy danh sách mã giảm giá với phân trang và lọc
// @Tags Coupon
// @Accept json
// @Produce json
// @Param page query int false "Số trang" default(1)
// @Param limit query int false "Số item mỗi trang" default(10)
// @Param active query bool false "Lọc theo trạng thái active"
// @Param code query string false "Tìm kiếm theo mã"
// @Success 200 {object} dto.CouponListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/coupons [get]
func (h *CouponHandler) GetCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")
	code := c.Query("code")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query với filters
	query := `
		SELECT c.id, c.code, c.description, c.discount_type, c.discount_value,
		       c.min_order_amount, c.max_uses, c.used_count, c.is_active,
		       c.valid_from, c.valid_until, c.created_at, c.updated_at,
		       COUNT(*) OVER() as total_count
		FROM coupons c
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if active != "" {
		if active == "true" {
			query += fmt.Sprintf(" AND c.is_active = $%d", argIndex)
			args = append(args, true)
			argIndex++
		} else if active == "false" {
			query += fmt.Sprintf(" AND c.is_active = $%d", argIndex)
			args = append(args, false)
			argIndex++
		}
	}

	if code != "" {
		query += fmt.Sprintf(" AND c.code ILIKE $%d", argIndex)
		args = append(args, "%"+code+"%")
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch coupons",
		})
		return
	}
	defer rows.Close()

	var coupons []dto.CouponDTO
	var totalCount int

	for rows.Next() {
		var coupon dto.CouponDTO
		var description sql.NullString
		var minOrderAmount sql.NullFloat64
		var maxUses sql.NullInt64
		var validUntil sql.NullTime

		err := rows.Scan(
			&coupon.ID, &coupon.Code, &description, &coupon.DiscountType,
			&coupon.DiscountValue, &minOrderAmount, &maxUses, &coupon.UsedCount,
			&coupon.IsActive, &coupon.ValidFrom, &validUntil,
			&coupon.CreatedAt, &coupon.UpdatedAt, &totalCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Scan error",
				Message: "Failed to parse coupon data",
			})
			return
		}

		if description.Valid {
			coupon.Description = &description.String
		}
		if minOrderAmount.Valid {
			coupon.MinOrderAmount = &minOrderAmount.Float64
		}
		if maxUses.Valid {
			maxUsesInt := int(maxUses.Int64)
			coupon.MaxUses = &maxUsesInt
		}
		if validUntil.Valid {
			coupon.ValidUntil = &validUntil.Time
		}

		coupons = append(coupons, coupon)
	}

	// Tính toán pagination
	totalPages := (totalCount + limit - 1) / limit

	response := dto.CouponListResponse{
		Data: coupons,
		Pagination: dto.PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCoupon godoc
// @Summary Lấy thông tin mã giảm giá theo ID
// @Description Lấy thông tin chi tiết mã giảm giá theo ID
// @Tags Coupon
// @Accept json
// @Produce json
// @Param id path string true "ID mã giảm giá"
// @Success 200 {object} dto.CouponDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/coupons/{id} [get]
func (h *CouponHandler) GetCoupon(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid coupon ID format",
		})
		return
	}

	query := `
		SELECT id, code, description, discount_type, discount_value,
		       min_order_amount, max_uses, used_count, is_active,
		       valid_from, valid_until, created_at, updated_at
		FROM coupons 
		WHERE id = $1`

	var coupon dto.CouponDTO
	var description sql.NullString
	var minOrderAmount sql.NullFloat64
	var maxUses sql.NullInt64
	var validUntil sql.NullTime

	err := h.db.QueryRow(query, id).Scan(
		&coupon.ID, &coupon.Code, &description, &coupon.DiscountType,
		&coupon.DiscountValue, &minOrderAmount, &maxUses, &coupon.UsedCount,
		&coupon.IsActive, &coupon.ValidFrom, &validUntil,
		&coupon.CreatedAt, &coupon.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not found",
				Message: "Coupon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to fetch coupon",
		})
		return
	}

	if description.Valid {
		coupon.Description = &description.String
	}
	if minOrderAmount.Valid {
		coupon.MinOrderAmount = &minOrderAmount.Float64
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int64)
		coupon.MaxUses = &maxUsesInt
	}
	if validUntil.Valid {
		coupon.ValidUntil = &validUntil.Time
	}

	c.JSON(http.StatusOK, coupon)
}

// CreateCoupon godoc
// @Summary Tạo mã giảm giá mới
// @Description Tạo mã giảm giá mới
// @Tags Coupon
// @Accept json
// @Produce json
// @Param body body dto.CreateCouponRequest true "Thông tin mã giảm giá"
// @Success 201 {object} dto.CouponDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/coupons [post]
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req dto.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra code đã tồn tại chưa
	var codeExists bool
	h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM coupons WHERE code = $1)", req.Code).Scan(&codeExists)
	if codeExists {
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:   "Conflict",
			Message: "Coupon code already exists",
		})
		return
	}

	id := uuid.New().String()
	now := time.Now()

	// Set default values
	validFrom := now
	if req.ValidFrom != nil {
		validFrom = *req.ValidFrom
	}

	var description sql.NullString
	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: true}
	}

	var minOrderAmount sql.NullFloat64
	if req.MinOrderAmount != nil {
		minOrderAmount = sql.NullFloat64{Float64: *req.MinOrderAmount, Valid: true}
	}

	var maxUses sql.NullInt64
	if req.MaxUses != nil {
		maxUses = sql.NullInt64{Int64: int64(*req.MaxUses), Valid: true}
	}

	var validUntil sql.NullTime
	if req.ValidUntil != nil {
		validUntil = sql.NullTime{Time: *req.ValidUntil, Valid: true}
	}

	query := `
		INSERT INTO coupons (id, code, description, discount_type, discount_value, min_order_amount, 
		                   max_uses, used_count, is_active, valid_from, valid_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, code, description, discount_type, discount_value, min_order_amount, 
		          max_uses, used_count, is_active, valid_from, valid_until, created_at, updated_at`

	var coupon dto.CouponDTO

	err := h.db.QueryRow(query, id, req.Code, description, req.DiscountType, req.DiscountValue,
		minOrderAmount, maxUses, 0, true, validFrom, validUntil, now, now).Scan(
		&coupon.ID, &coupon.Code, &description, &coupon.DiscountType,
		&coupon.DiscountValue, &minOrderAmount, &maxUses, &coupon.UsedCount,
		&coupon.IsActive, &coupon.ValidFrom, &validUntil,
		&coupon.CreatedAt, &coupon.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create coupon",
		})
		return
	}

	if description.Valid {
		coupon.Description = &description.String
	}
	if minOrderAmount.Valid {
		coupon.MinOrderAmount = &minOrderAmount.Float64
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int64)
		coupon.MaxUses = &maxUsesInt
	}
	if validUntil.Valid {
		coupon.ValidUntil = &validUntil.Time
	}

	c.JSON(http.StatusCreated, coupon)
}

// ValidateCoupon godoc
// @Summary Validate mã giảm giá
// @Description Kiểm tra mã giảm giá có hợp lệ không và tính toán số tiền giảm
// @Tags Coupon
// @Accept json
// @Produce json
// @Param body body dto.ValidateCouponRequest true "Thông tin validate"
// @Success 200 {object} dto.ValidateCouponResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/coupons/validate [post]
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var req dto.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	query := `
		SELECT id, code, description, discount_type, discount_value,
		       min_order_amount, max_uses, used_count, is_active,
		       valid_from, valid_until, created_at, updated_at
		FROM coupons 
		WHERE code = $1`

	var coupon dto.CouponDTO
	var description sql.NullString
	var minOrderAmount sql.NullFloat64
	var maxUses sql.NullInt64
	var validUntil sql.NullTime

	err := h.db.QueryRow(query, req.Code).Scan(
		&coupon.ID, &coupon.Code, &description, &coupon.DiscountType,
		&coupon.DiscountValue, &minOrderAmount, &maxUses, &coupon.UsedCount,
		&coupon.IsActive, &coupon.ValidFrom, &validUntil,
		&coupon.CreatedAt, &coupon.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, dto.ValidateCouponResponse{
				IsValid: false,
				Message: "Coupon not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to validate coupon",
		})
		return
	}

	// Parse coupon data
	if description.Valid {
		coupon.Description = &description.String
	}
	if minOrderAmount.Valid {
		coupon.MinOrderAmount = &minOrderAmount.Float64
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int64)
		coupon.MaxUses = &maxUsesInt
	}
	if validUntil.Valid {
		coupon.ValidUntil = &validUntil.Time
	}

	// Validate coupon
	now := time.Now()
	response := dto.ValidateCouponResponse{
		Coupon: &coupon,
	}

	// Check if coupon is active
	if !coupon.IsActive {
		response.IsValid = false
		response.Message = "Coupon is inactive"
		c.JSON(http.StatusOK, response)
		return
	}

	// Check if coupon is within valid period
	if now.Before(coupon.ValidFrom) {
		response.IsValid = false
		response.Message = "Coupon is not yet valid"
		c.JSON(http.StatusOK, response)
		return
	}

	if coupon.ValidUntil != nil && now.After(*coupon.ValidUntil) {
		response.IsValid = false
		response.Message = "Coupon has expired"
		c.JSON(http.StatusOK, response)
		return
	}

	// Check usage limit
	if coupon.MaxUses != nil && coupon.UsedCount >= *coupon.MaxUses {
		response.IsValid = false
		response.Message = "Coupon usage limit exceeded"
		c.JSON(http.StatusOK, response)
		return
	}

	// Check minimum order amount
	if coupon.MinOrderAmount != nil && req.OrderAmount < *coupon.MinOrderAmount {
		response.IsValid = false
		response.Message = fmt.Sprintf("Minimum order amount is %.2f", *coupon.MinOrderAmount)
		c.JSON(http.StatusOK, response)
		return
	}

	// Calculate discount amount
	var discountAmount float64
	if coupon.DiscountType == "percentage" {
		discountAmount = req.OrderAmount * (coupon.DiscountValue / 100)
	} else { // fixed
		discountAmount = coupon.DiscountValue
		if discountAmount > req.OrderAmount {
			discountAmount = req.OrderAmount
		}
	}

	response.IsValid = true
	response.Message = "Coupon is valid"
	response.DiscountAmount = &discountAmount

	c.JSON(http.StatusOK, response)
}

// UpdateCoupon godoc
// @Summary Cập nhật mã giảm giá
// @Description Cập nhật thông tin mã giảm giá
// @Tags Coupon
// @Accept json
// @Produce json
// @Param id path string true "ID mã giảm giá"
// @Param body body dto.UpdateCouponRequest true "Thông tin cập nhật"
// @Success 200 {object} dto.CouponDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/coupons/{id} [put]
func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid coupon ID format",
		})
		return
	}

	var req dto.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid input",
			Message: err.Error(),
		})
		return
	}

	// Kiểm tra coupon tồn tại
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM coupons WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Coupon not found",
		})
		return
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Code != nil {
		// Kiểm tra code mới có trùng không
		var codeExists bool
		h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM coupons WHERE code = $1 AND id != $2)", *req.Code, id).Scan(&codeExists)
		if codeExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Coupon code already exists",
			})
			return
		}
		setParts = append(setParts, fmt.Sprintf("code = $%d", argIndex))
		args = append(args, *req.Code)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.DiscountType != nil {
		setParts = append(setParts, fmt.Sprintf("discount_type = $%d", argIndex))
		args = append(args, *req.DiscountType)
		argIndex++
	}

	if req.DiscountValue != nil {
		setParts = append(setParts, fmt.Sprintf("discount_value = $%d", argIndex))
		args = append(args, *req.DiscountValue)
		argIndex++
	}

	if req.MinOrderAmount != nil {
		setParts = append(setParts, fmt.Sprintf("min_order_amount = $%d", argIndex))
		args = append(args, *req.MinOrderAmount)
		argIndex++
	}

	if req.MaxUses != nil {
		setParts = append(setParts, fmt.Sprintf("max_uses = $%d", argIndex))
		args = append(args, *req.MaxUses)
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if req.ValidFrom != nil {
		setParts = append(setParts, fmt.Sprintf("valid_from = $%d", argIndex))
		args = append(args, *req.ValidFrom)
		argIndex++
	}

	if req.ValidUntil != nil {
		setParts = append(setParts, fmt.Sprintf("valid_until = $%d", argIndex))
		args = append(args, *req.ValidUntil)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No updates",
			Message: "No fields to update",
		})
		return
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf(`
		UPDATE coupons SET %s 
		WHERE id = $%d
		RETURNING id, code, description, discount_type, discount_value, min_order_amount, 
		          max_uses, used_count, is_active, valid_from, valid_until, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]),
		argIndex,
	)

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE coupons SET %s, %s 
			WHERE id = $%d
			RETURNING id, code, description, discount_type, discount_value, min_order_amount, 
			          max_uses, used_count, is_active, valid_from, valid_until, created_at, updated_at`,
			setParts[0], setParts[i],
			argIndex,
		)
	}

	args = append(args, id)

	var coupon dto.CouponDTO
	var description sql.NullString
	var minOrderAmount sql.NullFloat64
	var maxUses sql.NullInt64
	var validUntil sql.NullTime

	err = h.db.QueryRow(query, args...).Scan(
		&coupon.ID, &coupon.Code, &description, &coupon.DiscountType,
		&coupon.DiscountValue, &minOrderAmount, &maxUses, &coupon.UsedCount,
		&coupon.IsActive, &coupon.ValidFrom, &validUntil,
		&coupon.CreatedAt, &coupon.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update coupon",
		})
		return
	}

	if description.Valid {
		coupon.Description = &description.String
	}
	if minOrderAmount.Valid {
		coupon.MinOrderAmount = &minOrderAmount.Float64
	}
	if maxUses.Valid {
		maxUsesInt := int(maxUses.Int64)
		coupon.MaxUses = &maxUsesInt
	}
	if validUntil.Valid {
		coupon.ValidUntil = &validUntil.Time
	}

	c.JSON(http.StatusOK, coupon)
}

// DeleteCoupon godoc
// @Summary Xóa mã giảm giá
// @Description Xóa mã giảm giá theo ID
// @Tags Coupon
// @Accept json
// @Produce json
// @Param id path string true "ID mã giảm giá"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/coupons/{id} [delete]
func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid ID",
			Message: "Invalid coupon ID format",
		})
		return
	}

	result, err := h.db.Exec("DELETE FROM coupons WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to delete coupon",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not found",
			Message: "Coupon not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
