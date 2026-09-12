package model

import "time"

// Conversation is a private chat thread between buyer and seller for a product.
type Conversation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"uniqueIndex:uniq_conversation_triple,priority:1;not null" json:"product_id"`
	BuyerID   uint      `gorm:"uniqueIndex:uniq_conversation_triple,priority:2;not null" json:"buyer_id"`
	SellerID  uint      `gorm:"uniqueIndex:uniq_conversation_triple,priority:3;not null" json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Message is one chat message inside a conversation.
type Message struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ConversationID uint      `gorm:"index;not null" json:"conversation_id"`
	SenderID       uint      `gorm:"not null" json:"sender_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	Read           bool      `gorm:"default:false" json:"read"`
	CreatedAt      time.Time `json:"created_at"`
}
