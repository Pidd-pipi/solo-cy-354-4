package repository

import (
	"context"

	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

// TradeOrderRepository persists trade order rows. It only reads and writes;
// every allowed status transition and timestamp column is decided by the
// caller (the trade state rules), never here.
type TradeOrderRepository struct {
	db *gorm.DB
}

// NewTradeOrderRepository builds a TradeOrderRepository.
func NewTradeOrderRepository(db *gorm.DB) *TradeOrderRepository {
	return &TradeOrderRepository{db: db}
}

// Transaction runs fn inside a database transaction for cross-repository writes.
func (r *TradeOrderRepository) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return Transaction(ctx, r.db, fn)
}

// Create inserts a new trade order. The database active-order unique key
// guarantees that two concurrent creates for the same product and buyer
// cannot both succeed; the losing insert comes back as util.ErrConflict.
func (r *TradeOrderRepository) Create(ctx context.Context, o *model.TradeOrder) error {
	if err := db(ctx, r.db).Create(o).Error; err != nil {
		return mapTradeOrderCreateError(err)
	}
	return nil
}

// FindByID returns a trade order by id.
func (r *TradeOrderRepository) FindByID(ctx context.Context, id uint) (*model.TradeOrder, error) {
	var o model.TradeOrder
	err := db(ctx, r.db).First(&o, id).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &o, nil
}

// FindByProductBuyerStatuses returns a buyer's order for a product whose
// status is in the given list. Callers pass the statuses they consider a
// match; the repository holds no opinion on which statuses those are.
func (r *TradeOrderRepository) FindByProductBuyerStatuses(ctx context.Context, productID, buyerID uint, statuses []string) (*model.TradeOrder, error) {
	var o model.TradeOrder
	err := db(ctx, r.db).
		Where("product_id = ? AND buyer_id = ? AND status IN ?", productID, buyerID, statuses).
		First(&o).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &o, nil
}

// ListByUser returns orders where the user is buyer or seller.
func (r *TradeOrderRepository) ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]model.TradeOrder, int64, error) {
	q := db(ctx, r.db).Model(&model.TradeOrder{}).Where("buyer_id = ? OR seller_id = ?", userID, userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.TradeOrder
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Transition applies an atomic compare-and-set on the order status: it only
// succeeds while the row still has expectStatus, then writes toStatus and the
// caller-provided timestamp columns. A 0 RowsAffected means the row vanished
// (ErrNotFound) or the current status no longer matches expectStatus
// (ErrConflict); state validation itself lives in the service layer.
func (r *TradeOrderRepository) Transition(ctx context.Context, id uint, expectStatus, toStatus string, timestampColumns map[string]interface{}) error {
	updates := mergeStatusUpdates(toStatus, timestampColumns)
	res := db(ctx, r.db).Model(&model.TradeOrder{}).
		Where("id = ? AND status = ?", id, expectStatus).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		if err := normalizeError(db(ctx, r.db).Select("id").First(&model.TradeOrder{}, id).Error); err != nil {
			return err
		}
		return util.ErrConflict
	}
	return nil
}

func mergeStatusUpdates(toStatus string, timestampColumns map[string]interface{}) map[string]interface{} {
	updates := make(map[string]interface{}, len(timestampColumns)+1)
	for col, val := range timestampColumns {
		updates[col] = val
	}
	updates["status"] = toStatus
	return updates
}
