package repository

import (
	"context"

	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

// ProductRepository persists product rows.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository builds a ProductRepository.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product.
func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	return db(ctx, r.db).Create(p).Error
}

// FindByID returns a product by id.
func (r *ProductRepository) FindByID(ctx context.Context, id uint) (*model.Product, error) {
	var p model.Product
	err := db(ctx, r.db).First(&p, id).Error
	if err != nil {
		return nil, normalizeError(err)
	}
	return &p, nil
}

// List filters products by category/campus/keyword/status with pagination.
func (r *ProductRepository) List(ctx context.Context, category, campus, keyword, status string, page, pageSize int) ([]model.Product, int64, error) {
	q := db(ctx, r.db).Model(&model.Product{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if campus != "" {
		q = q.Where("campus = ?", campus)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Product
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpdateStatus sets the product status.
func (r *ProductRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	res := db(ctx, r.db).Model(&model.Product{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return util.ErrNotFound
	}
	return nil
}

// Count returns the total product count.
func (r *ProductRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := db(ctx, r.db).Model(&model.Product{}).Count(&n).Error
	return n, err
}
