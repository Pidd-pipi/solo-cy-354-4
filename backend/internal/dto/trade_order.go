package dto

// CreateTradeOrderRequest creates a purchase intent for a product.
type CreateTradeOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}
