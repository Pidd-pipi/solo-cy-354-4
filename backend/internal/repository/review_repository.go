package repository

import (
	"context"

	"github.com/lp/campus-market/internal/model"
	"gorm.io/gorm"
)

// ReviewRepository persists review rows.
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository builds a ReviewRepository.
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Transaction runs fn inside a database transaction for cross-repository writes.
func (r *ReviewRepository) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return Transaction(ctx, r.db, fn)
}

// Create inserts a new review.
func (r *ReviewRepository) Create(ctx context.Context, rv *model.Review) error {
	return db(ctx, r.db).Create(rv).Error
}

// FindByTradeAndReviewer returns an existing review for a trade by the reviewer.
func (r *ReviewRepository) FindByTradeAndReviewer(ctx context.Context, tradeID, reviewerID uint) (*model.Review, error) {
	var rv model.Review
	err := db(ctx, r.db).Where("trade_id = ? AND reviewer_id = ?", tradeID, reviewerID).First(&rv).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &rv, nil
}

// ListByReviewee returns reviews received by a user.
func (r *ReviewRepository) ListByReviewee(ctx context.Context, userID uint) ([]model.Review, error) {
	var items []model.Review
	err := db(ctx, r.db).Where("reviewee_id = ?", userID).Order("created_at DESC").Find(&items).Error
	return items, err
}
