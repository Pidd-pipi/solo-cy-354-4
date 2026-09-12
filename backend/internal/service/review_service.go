package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/repository"
	"github.com/lp/campus-market/internal/util"
)

// ReviewService manages post-trade reviews and credit scoring.
type ReviewService struct {
	reviews *repository.ReviewRepository
	orders  *repository.TradeOrderRepository
	users   *repository.UserRepository
	logger  *slog.Logger
}

// NewReviewService wires the review service dependencies.
func NewReviewService(reviews *repository.ReviewRepository, orders *repository.TradeOrderRepository, users *repository.UserRepository, logger *slog.Logger) *ReviewService {
	return &ReviewService{reviews: reviews, orders: orders, users: users, logger: logger}
}

// Create leaves a review for a completed trade and updates the reviewee credit.
func (s *ReviewService) Create(ctx context.Context, reviewer *model.User, req *dto.CreateReviewRequest) (*model.Review, error) {
	order, err := s.orders.FindByID(ctx, req.TradeID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("review[trade=%d] order lookup: %w", req.TradeID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if order.Status != constants.TradeStatusCompleted {
		return nil, util.NewAppError(409, constants.CodeConflict, "交易未完成不可评价", nil)
	}
	if order.BuyerID != reviewer.ID && order.SellerID != reviewer.ID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgForbidden, nil)
	}
	revieweeID := order.SellerID
	if reviewer.ID == order.SellerID {
		revieweeID = order.BuyerID
	}
	rv := &model.Review{
		TradeID: req.TradeID, ReviewerID: reviewer.ID, RevieweeID: revieweeID,
		Rating: req.Rating, Content: req.Content,
	}
	delta := util.CreditDelta(req.Rating)
	if err := s.reviews.Transaction(ctx, func(txCtx context.Context) error {
		if _, err := s.reviews.FindByTradeAndReviewer(txCtx, req.TradeID, reviewer.ID); err == nil {
			return util.ErrConflict
		} else if !errors.Is(err, util.ErrNotFound) {
			return err
		}
		if err := s.reviews.Create(txCtx, rv); err != nil {
			return err
		}
		if err := s.users.AddCredit(txCtx, revieweeID, delta); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, util.ErrConflict) {
			return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgAlreadyReviewed, nil)
		}
		s.logger.Error(fmt.Sprintf(constants.LogReviewCreateFailed, req.TradeID, err))
		return nil, util.WrapAppError(fmt.Errorf("review[trade=%d] create: %w", req.TradeID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogReviewCreateSuccess, rv.ID, req.TradeID, req.Rating))
	if delta != 0 {
		s.logger.Info(fmt.Sprintf(constants.LogCreditUpdateSuccess, revieweeID, delta))
	}
	return rv, nil
}

// ListByReviewee returns reviews received by a user.
func (s *ReviewService) ListByReviewee(ctx context.Context, userID uint) ([]model.Review, error) {
	items, err := s.reviews.ListByReviewee(ctx, userID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("review[reviewee=%d] list: %w", userID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	return items, nil
}

var _ = errors.Is
