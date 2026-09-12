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
	"github.com/lp/campus-market/internal/service/tradestate"
	"github.com/lp/campus-market/internal/util"
)

// OrderRepository is the read/write contract for trade order rows. It exposes
// only persistence; all status rules live in the tradestate package.
type OrderRepository interface {
	Create(ctx context.Context, o *model.TradeOrder) error
	FindByID(ctx context.Context, id uint) (*model.TradeOrder, error)
	FindByProductBuyerStatuses(ctx context.Context, productID, buyerID uint, statuses []string) (*model.TradeOrder, error)
	ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]model.TradeOrder, int64, error)
	Transition(ctx context.Context, id uint, expectStatus, toStatus string, timestampColumns map[string]interface{}) error
	Transaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

// TradeOrderService manages purchase intents, confirmations and completion.
// Every lifecycle action follows the same path: load the order, ask the
// tradestate rules whether the actor may perform the action in the current
// status, then persist the approved transition atomically.
type TradeOrderService struct {
	orders   OrderRepository
	products ProductRepository
	logger   *slog.Logger
}

// NewTradeOrderService wires the trade order service dependencies.
func NewTradeOrderService(orders OrderRepository, products ProductRepository, logger *slog.Logger) *TradeOrderService {
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
	existing, err := s.orders.FindByProductBuyerStatuses(ctx, req.ProductID, buyer.ID, tradestate.ActiveStatuses())
	if err != nil && !errors.Is(err, util.ErrNotFound) {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=? product=%d buyer=%d] duplicate order lookup: %w", req.ProductID, buyer.ID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	if existing != nil {
		return nil, util.NewAppError(409, constants.CodeConflict, "您已对该商品下单", nil)
	}
	order := &model.TradeOrder{
		ProductID: req.ProductID, BuyerID: buyer.ID, SellerID: product.SellerID,
		Status: tradestate.Initial(),
	}
	if err := s.orders.Create(ctx, order); err != nil {
		// The database active-order unique key is the race-proof backstop:
		// a concurrent request that passed the lookup first wins the one
		// active order and this insert is rejected as a duplicate.
		if errors.Is(err, util.ErrConflict) {
			return nil, util.NewAppError(409, constants.CodeConflict, "您已对该商品下单", nil)
		}
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=? product=%d buyer=%d] create: %w", req.ProductID, buyer.ID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
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
	order, err := s.load(ctx, orderID, "buyer confirm find")
	if err != nil {
		return nil, err
	}
	rule, err := tradestate.Guard(order, tradestate.ActionBuyerConfirm, userID)
	if err != nil {
		return nil, guardError(orderID, err)
	}
	if err := s.orders.Transition(ctx, orderID, rule.From, rule.To, rule.ColumnUpdates(time.Now())); err != nil {
		return nil, s.persistError(orderID, "buyer confirm", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderBuyerConfirmSuccess, orderID))
	order.Status = rule.To
	return order, nil
}

// SellerConfirm completes the order and marks the product sold.
func (s *TradeOrderService) SellerConfirm(ctx context.Context, userID, orderID uint) (*model.TradeOrder, error) {
	order, err := s.load(ctx, orderID, "seller confirm find")
	if err != nil {
		return nil, err
	}
	rule, err := tradestate.Guard(order, tradestate.ActionSellerConfirm, userID)
	if err != nil {
		return nil, guardError(orderID, err)
	}
	now := time.Now()
	if err := s.orders.Transaction(ctx, func(txCtx context.Context) error {
		if err := s.orders.Transition(txCtx, orderID, rule.From, rule.To, rule.ColumnUpdates(now)); err != nil {
			return err
		}
		return s.products.UpdateStatus(txCtx, order.ProductID, rule.ProductOnComplete)
	}); err != nil {
		if errors.Is(err, util.ErrConflict) {
			return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
		}
		s.logger.Error(fmt.Sprintf(constants.LogTradeOrderCompleteFailed, orderID, err))
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] seller confirm: %w", orderID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderCompleteSuccess, orderID, order.ProductID))
	order.Status = rule.To
	return order, nil
}

// Cancel cancels an order for the buyer or seller.
func (s *TradeOrderService) Cancel(ctx context.Context, userID, orderID uint) (*model.TradeOrder, error) {
	order, err := s.load(ctx, orderID, "cancel find")
	if err != nil {
		return nil, err
	}
	rule, err := tradestate.Guard(order, tradestate.ActionCancel, userID)
	if err != nil {
		return nil, guardError(orderID, err)
	}
	if err := s.orders.Transition(ctx, orderID, rule.From, rule.To, rule.ColumnUpdates(time.Now())); err != nil {
		return nil, s.persistError(orderID, "cancel", err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTradeOrderCancelSuccess, orderID))
	order.Status = rule.To
	return order, nil
}

// load fetches the order and converts a missing row into the standard 404.
func (s *TradeOrderService) load(ctx context.Context, orderID uint, stage string) (*model.TradeOrder, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("trade_order[id=%d] %s: %w", orderID, stage, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	return order, nil
}

// guardError maps a state-rule rejection onto the same response every action
// returns: wrong actor -> 403, status not allowed -> 409.
func guardError(orderID uint, err error) error {
	switch {
	case errors.Is(err, tradestate.ErrIllegalRole):
		return util.NewAppError(403, constants.CodeForbidden, constants.MsgNotParticipant, nil)
	case errors.Is(err, tradestate.ErrIllegalTransition):
		return util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
	default:
		return util.WrapAppError(fmt.Errorf("trade_order[id=%d] guard: %w", orderID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
}

// persistError converts a failed transition write: a concurrent status change
// is a 409, a vanished row a 404, anything else a 500.
func (s *TradeOrderService) persistError(orderID uint, stage string, err error) error {
	switch {
	case errors.Is(err, util.ErrConflict):
		return util.NewAppError(409, constants.CodeConflict, constants.MsgTradeStatusInvalid, nil)
	case errors.Is(err, util.ErrNotFound):
		return util.NewAppError(404, constants.CodeNotFound, constants.MsgNotFound, nil)
	default:
		return util.WrapAppError(fmt.Errorf("trade_order[id=%d] %s: %w", orderID, stage, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
}
