package repository

import (
	"context"

	"github.com/lp/campus-market/internal/model"
	"gorm.io/gorm"
)

// ConversationRepository persists conversation and message rows.
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository builds a ConversationRepository.
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// FindExisting returns a conversation between buyer and seller for a product.
func (r *ConversationRepository) FindExisting(ctx context.Context, productID, buyerID, sellerID uint) (*model.Conversation, error) {
	var c model.Conversation
	err := db(ctx, r.db).Where("product_id = ? AND buyer_id = ? AND seller_id = ?", productID, buyerID, sellerID).First(&c).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &c, nil
}

// Create inserts a new conversation.
func (r *ConversationRepository) Create(ctx context.Context, c *model.Conversation) error {
	return db(ctx, r.db).Create(c).Error
}

// FindByID returns a conversation by id.
func (r *ConversationRepository) FindByID(ctx context.Context, id uint) (*model.Conversation, error) {
	var c model.Conversation
	err := db(ctx, r.db).First(&c, id).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &c, nil
}

// ListByUser returns conversations where the user is buyer or seller.
func (r *ConversationRepository) ListByUser(ctx context.Context, userID uint) ([]model.Conversation, error) {
	var items []model.Conversation
	err := db(ctx, r.db).Where("buyer_id = ? OR seller_id = ?", userID, userID).Order("created_at DESC").Find(&items).Error
	return items, err
}

// CreateMessage inserts a chat message.
func (r *ConversationRepository) CreateMessage(ctx context.Context, m *model.Message) error {
	return db(ctx, r.db).Create(m).Error
}

// ListMessages returns messages of a conversation ordered ascending.
func (r *ConversationRepository) ListMessages(ctx context.Context, conversationID uint) ([]model.Message, error) {
	var items []model.Message
	err := db(ctx, r.db).Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&items).Error
	return items, err
}
