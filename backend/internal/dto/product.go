package dto

// CreateProductRequest is the payload for publishing a second-hand item.
type CreateProductRequest struct {
	Title         string  `json:"title" binding:"required,min=2,max=64"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	Category      string  `json:"category" binding:"required"`
	Condition     string  `json:"condition" binding:"required"`
	Campus        string  `json:"campus" binding:"required,max=64"`
	TradeLocation string  `json:"trade_location" binding:"required,max=128"`
	Images        string  `json:"images"`
}

// ListProductQuery adds filters to pagination.
type ListProductQuery struct {
	PageQuery
	Category string `form:"category"`
	Campus   string `form:"campus"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
}
