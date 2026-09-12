package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/repository"
	"github.com/lp/campus-market/internal/util"
)

// ConversationService manages price-negotiation chats between buyer and seller.
type ConversationService struct {
	convs *repository.ConversationRepository
	logger *slog.Logger
}

// NewConversationService wires the conversation service dependencies.
func NewConversationService(convs *repository.ConversationRepository, logger *slog.Logger) *ConversationService {
	return &ConversationService{convs: convs, logger: logger}
}

// Create starts or reuses a chat thread for a product.
func (s *ConversationService) Create(ctx context.Context, buyer *model.User, sellerID uint, req *dto.CreateConversationRequest) (*model.Conversation, error) {
	existing, err := s.convs.FindExisting(ctx, req.ProductID, buyer.ID, sellerID)
	if err == nil && existing != nil {
		return existing, nil
	}
	conv := &model.Conversation{ProductID: req.ProductID, BuyerID: buyer.ID, SellerID: sellerID}
	if err := s.convs.Create(ctx, conv); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("conversation[buyer=%d] create: %w", buyer.ID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogConversationCreateSuccess, conv.ID, req.ProductID))
	return conv, nil
}

// ListMy returns conversations where the user participates.
func (s *ConversationService) ListMy(ctx context.Context, userID uint) ([]model.Conversation, error) {
	items, err := s.convs.ListByUser(ctx, userID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("conversation[user=%d] list: %w", userID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	return items, nil
}

// SendMessage appends a message to a conversation.
func (s *ConversationService) SendMessage(ctx context.Context, senderID uint, convID uint, content string) (*model.Message, error) {
	conv, err := s.convs.FindByID(ctx, convID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("conversation[id=%d] send find: %w", convID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if conv.BuyerID != senderID && conv.SellerID != senderID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgForbidden, nil)
	}
	msg := &model.Message{ConversationID: convID, SenderID: senderID, Content: content, Read: false}
	if err := s.convs.CreateMessage(ctx, msg); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("conversation[id=%d] send: %w", convID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMessageSendSuccess, convID, senderID))
	return msg, nil
}

// ListMessages returns the chat history of a conversation.
func (s *ConversationService) ListMessages(ctx context.Context, userID, convID uint) ([]model.Message, error) {
	conv, err := s.convs.FindByID(ctx, convID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("conversation[id=%d] messages find: %w", convID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if conv.BuyerID != userID && conv.SellerID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgForbidden, nil)
	}
	items, err := s.convs.ListMessages(ctx, convID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("conversation[id=%d] messages: %w", convID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	return items, nil
}
