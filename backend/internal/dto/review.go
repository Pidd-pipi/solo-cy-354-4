package dto

// CreateReviewRequest is the payload for rating a completed trade.
type CreateReviewRequest struct {
	TradeID uint   `json:"trade_id" binding:"required"`
	Rating  string `json:"rating" binding:"required,oneof=good medium bad"`
	Content string `json:"content" binding:"max=500"`
}
