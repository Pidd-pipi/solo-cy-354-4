package dto

// CreateConversationRequest starts a chat thread for a product.
type CreateConversationRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// SendMessageRequest is the payload for sending a chat message.
type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}
