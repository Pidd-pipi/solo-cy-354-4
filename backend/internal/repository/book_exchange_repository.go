package repository

import (
	"context"

	"github.com/lp/campus-market/internal/model"
	"gorm.io/gorm"
)

// BookExchangeRepository persists book swap rows.
type BookExchangeRepository struct {
	db *gorm.DB
}

// NewBookExchangeRepository builds a BookExchangeRepository.
func NewBookExchangeRepository(db *gorm.DB) *BookExchangeRepository {
	return &BookExchangeRepository{db: db}
}

// Transaction runs fn inside a database transaction for cross-repository writes.
func (r *BookExchangeRepository) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return Transaction(ctx, r.db, fn)
}

// Create inserts a new book exchange.
func (r *BookExchangeRepository) Create(ctx context.Context, e *model.BookExchange) error {
	return db(ctx, r.db).Create(e).Error
}

// FindByID returns a book exchange by id.
func (r *BookExchangeRepository) FindByID(ctx context.Context, id uint) (*model.BookExchange, error) {
	var e model.BookExchange
	err := db(ctx, r.db).First(&e, id).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &e, nil
}

// List returns open exchanges with pagination.
func (r *BookExchangeRepository) List(ctx context.Context, page, pageSize int) ([]model.BookExchange, int64, error) {
	q := db(ctx, r.db).Model(&model.BookExchange{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.BookExchange
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// FindMatch returns an open exchange whose want matches another's offer and vice versa.
func (r *BookExchangeRepository) FindMatch(ctx context.Context, offer, want string, excludeID uint) (*model.BookExchange, error) {
	var e model.BookExchange
	err := db(ctx, r.db).
		Where("status = ? AND offer_book = ? AND want_book = ? AND id <> ?", "open", want, offer, excludeID).
		First(&e).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &e, nil
}

// UpdateStatus sets the exchange status.
func (r *BookExchangeRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return db(ctx, r.db).Model(&model.BookExchange{}).Where("id = ?", id).Update("status", status).Error
}

// MarkMatched links the exchange to a matched partner.
func (r *BookExchangeRepository) MarkMatched(ctx context.Context, id, matchedID uint) error {
	return db(ctx, r.db).Model(&model.BookExchange{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "matched", "matched_id": matchedID}).Error
}
