package dto

import "time"

// CouponDTO - DTO cho mã giảm giá
type CouponDTO struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	Description    *string    `json:"description,omitempty"`
	DiscountType   string     `json:"discount_type"` // 'percentage' hoặc 'fixed'
	DiscountValue  float64    `json:"discount_value"`
	MinOrderAmount *float64   `json:"min_order_amount,omitempty"`
	MaxUses        *int       `json:"max_uses,omitempty"`
	UsedCount      int        `json:"used_count"`
	IsActive       bool       `json:"is_active"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateCouponRequest - Request tạo mã giảm giá
type CreateCouponRequest struct {
	Code           string     `json:"code" binding:"required,max=50"`
	Description    *string    `json:"description,omitempty"`
	DiscountType   string     `json:"discount_type" binding:"required,oneof=percentage fixed"`
	DiscountValue  float64    `json:"discount_value" binding:"required,gt=0"`
	MinOrderAmount *float64   `json:"min_order_amount,omitempty" binding:"omitempty,gte=0"`
	MaxUses        *int       `json:"max_uses,omitempty" binding:"omitempty,gt=0"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
}

// UpdateCouponRequest - Request cập nhật mã giảm giá
type UpdateCouponRequest struct {
	Code           *string    `json:"code,omitempty" binding:"omitempty,max=50"`
	Description    *string    `json:"description,omitempty"`
	DiscountType   *string    `json:"discount_type,omitempty" binding:"omitempty,oneof=percentage fixed"`
	DiscountValue  *float64   `json:"discount_value,omitempty" binding:"omitempty,gt=0"`
	MinOrderAmount *float64   `json:"min_order_amount,omitempty" binding:"omitempty,gte=0"`
	MaxUses        *int       `json:"max_uses,omitempty" binding:"omitempty,gt=0"`
	IsActive       *bool      `json:"is_active,omitempty"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
}

// CouponListResponse - Response danh sách mã giảm giá
type CouponListResponse struct {
	Data       []CouponDTO    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// ValidateCouponRequest - Request validate mã giảm giá
type ValidateCouponRequest struct {
	Code        string  `json:"code" binding:"required"`
	OrderAmount float64 `json:"order_amount" binding:"required,gt=0"`
}

// ValidateCouponResponse - Response validate mã giảm giá
type ValidateCouponResponse struct {
	IsValid       bool    `json:"is_valid"`
	Message       string  `json:"message"`
	DiscountAmount *float64 `json:"discount_amount,omitempty"`
	Coupon        *CouponDTO `json:"coupon,omitempty"`
}
