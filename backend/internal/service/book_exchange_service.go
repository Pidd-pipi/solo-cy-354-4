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

// BookExchangeService manages book swap posts and matching.
type BookExchangeService struct {
	exchanges *repository.BookExchangeRepository
	logger    *slog.Logger
}

// NewBookExchangeService wires the book exchange service dependencies.
func NewBookExchangeService(exchanges *repository.BookExchangeRepository, logger *slog.Logger) *BookExchangeService {
	return &BookExchangeService{exchanges: exchanges, logger: logger}
}

// Create publishes a book exchange request and tries to match it.
func (s *BookExchangeService) Create(ctx context.Context, userID uint, req *dto.CreateBookExchangeRequest) (*model.BookExchange, error) {
	e := &model.BookExchange{
		UserID: userID, OfferBook: req.OfferBook, WantBook: req.WantBook,
		Description: req.Description, Status: "open",
	}
	if err := s.exchanges.Transaction(ctx, func(txCtx context.Context) error {
		if err := s.exchanges.Create(txCtx, e); err != nil {
			return err
		}
		match, err := s.exchanges.FindMatch(txCtx, req.OfferBook, req.WantBook, e.ID)
		if err != nil && !errors.Is(err, util.ErrNotFound) {
			return err
		}
		if match == nil {
			return nil
		}
		if err := s.exchanges.MarkMatched(txCtx, e.ID, match.ID); err != nil {
			return err
		}
		if err := s.exchanges.MarkMatched(txCtx, match.ID, e.ID); err != nil {
			return err
		}
		e.Status = "matched"
		e.MatchedID = &match.ID
		return nil
	}); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("book_exchange[user=%d] create: %w", userID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBookExchangePublishSuccess, e.ID, userID))
	if e.Status == "matched" {
		s.logger.Info(fmt.Sprintf(constants.LogBookExchangeMatchSuccess, e.ID, *e.MatchedID))
	}
	return e, nil
}

// List returns exchanges with pagination.
func (s *BookExchangeService) List(ctx context.Context, q *dto.PageQuery) (*dto.PageResult, error) {
	q.Normalize()
	items, total, err := s.exchanges.List(ctx, q.Page, q.PageSize)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("book_exchange list: %w", err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	return &dto.PageResult{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Close closes an open exchange owned by the user.
func (s *BookExchangeService) Close(ctx context.Context, userID, exchangeID uint) (*model.BookExchange, error) {
	e, err := s.exchanges.FindByID(ctx, exchangeID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("book_exchange[id=%d] close find: %w", exchangeID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if e.UserID != userID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgForbidden, nil)
	}
	if e.Status == "closed" {
		return nil, util.NewAppError(409, constants.CodeConflict, constants.MsgExchangeClosed, nil)
	}
	if err := s.exchanges.UpdateStatus(ctx, exchangeID, "closed"); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("book_exchange[id=%d] close: %w", exchangeID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogBookExchangeCloseSuccess, exchangeID))
	e.Status = "closed"
	return e, nil
}
