package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/repository"
	"github.com/lp/campus-market/internal/util"
)

// TradeOrderService manages purchase intents, confirmations and completion.
type TradeOrderService struct {
	orders   *repository.TradeOrderRepository
	products *repository.ProductRepository
	logger   *slog.Logger
}

// NewTradeOrderService wires the trade order service dependencies.
func NewTradeOrderService(orders *repository.TradeOrderRepository, products *repository.ProductRepository, logger *slog.Logger) *TradeOrderService {
	return &TradeOrderService{orders: orders, products: products, logger: logger}
}

// Create creates a pending trade order for an on-sale product.
func (s *TradeOrderService) Create(ctx context.Context, buyer *model.User, req *dto.CreateTradeOrderRequest) (*model.TradeOrder, error) {
	product, err := s.products.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[buyer=%d] product lookup: %w", buyer.ID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if product.SellerID == buyer.ID {
		return nil, util.NewAppError(400, constants.CodeBadRequest, "不能购买自己的商品", nil)
	}
	if product.Status != constants.ProductStatusOnSale {
		return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgProductNotOnSale, nil)
	}
	if existing, err := s.orders.FindByProductAndBuyer(ctx, req.ProductID, buyer.ID); err == nil && existing != nil {
		return nil, util.NewAppError(409, constants.CodeConflict, "您已对该商品下单", nil)
	}
	order := &model.TradeOrder{
		ProductID: req.ProductID, BuyerID: buyer.ID, SellerID: product.SellerID,
		Status: constants.TradeStatusPending,
	}
	if err := s.orders.Create(ctx, order); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[buyer=%d] create: %w", buyer.ID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderCreateSuccess, order.ID, req.ProductID))
	return order, nil
}

// ListMy returns the orders where the user participates.
func (s *TradeOrderService) ListMy(ctx context.Context, userID uint, q *dto.PageQuery) (*dto.PageResult, error) {
	q.Normalize()
	items, total, err := s.orders.ListByUser(ctx, userID, q.Page, q.PageSize)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[user=%d] list: %w", userID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	return &dto.PageResult{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// BuyerConfirm marks the order confirmed by the buyer.
func (s *TradeOrderService) BuyerConfirm(ctx context.Context, userID, orderID uint) (*model.TradeOrder, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] buyer confirm find: %w", orderID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if order.BuyerID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgNotParticipant, nil)
	}
	if order.Status != constants.TradeStatusPending {
		return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
	}
	now := time.Now()
	if err := s.orders.UpdateBuyerConfirmed(ctx, orderID, now); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] buyer confirm: %w", orderID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderBuyerConfirmSuccess, orderID))
	order.Status = constants.TradeStatusConfirmed
	return order, nil
}

// SellerConfirm completes the order and marks the product sold.
func (s *TradeOrderService) SellerConfirm(ctx context.Context, userID, orderID uint) (*model.TradeOrder, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] seller confirm find: %w", orderID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if order.SellerID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgNotParticipant, nil)
	}
	if order.Status != constants.TradeStatusConfirmed {
		return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
	}
	now := time.Now()
	if err := s.orders.Transaction(ctx, func(txCtx context.Context) error {
		if err := s.orders.UpdateSellerConfirmed(txCtx, orderID, now); err != nil {
			return err
		}
		if err := s.products.UpdateStatus(txCtx, order.ProductID, constants.ProductStatusSold); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, util.ErrConflict) {
			return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
		}
		s.logger.Error(fmt.Sprintf(constants.LogTradeOrderCompleteFailed, orderID, err))
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] seller confirm: %w", orderID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderCompleteSuccess, orderID, order.ProductID))
	order.Status = constants.TradeStatusCompleted
	return order, nil
}

// Cancel cancels a pending order.
func (s *TradeOrderService) Cancel(ctx context.Context, userID, orderID uint) (*model.TradeOrder, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] cancel find: %w", orderID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgNotParticipant, nil)
	}
	if order.Status != constants.TradeStatusPending {
		return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
	}
	if err := s.orders.UpdateStatus(ctx, orderID, constants.TradeStatusCancelled); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] cancel: %w", orderID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderCancelSuccess, orderID))
	order.Status = constants.TradeStatusCancelled
	return order, nil
}
