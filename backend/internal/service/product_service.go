package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/util"
)

// ProductRepository is the data access contract for product rows.
type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	FindByID(ctx context.Context, id uint) (*model.Product, error)
	List(ctx context.Context, category, campus, keyword, status string, page, pageSize int) ([]model.Product, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
	Count(ctx context.Context) (int64, error)
}

// ProductService manages second-hand product publishing and lifecycle.
type ProductService struct {
	products ProductRepository
	logger   *slog.Logger
}

// NewProductService wires the product service dependencies.
func NewProductService(products ProductRepository, logger *slog.Logger) *ProductService {
	return &ProductService{products: products, logger: logger}
}

// Create publishes a new product.
func (s *ProductService) Create(ctx context.Context, sellerID uint, req *dto.CreateProductRequest) (*model.Product, error) {
	if !constants.IsProductCategory(req.Category) {
		return nil, util.NewAppError(400, constants.CodeValidation, "商品分类不合法", nil)
	}
	p := &model.Product{
		SellerID: sellerID, Title: req.Title, Description: req.Description,
		Price: req.Price, Category: req.Category, Condition: req.Condition,
		Campus: req.Campus, TradeLocation: req.TradeLocation, Images: req.Images,
		Status: constants.ProductStatusOnSale,
	}
	if err := s.products.Create(ctx, p); err != nil {
		s.logger.Error(fmt.Sprintf(constants.LogProductPublishFailed, sellerID, req.Title, err))
		return nil, util.WrapAppError(fmt.Errorf("product[seller=%d] publish: %w", sellerID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogProductPublishSuccess, p.ID, p.Title))
	return p, nil
}

// Get returns one product.
func (s *ProductService) Get(ctx context.Context, id uint) (*model.Product, error) {
	p, err := s.products.FindByID(ctx, id)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("product[id=%d] get: %w", id, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	return p, nil
}

// List filters products.
func (s *ProductService) List(ctx context.Context, q *dto.ListProductQuery) (*dto.PageResult, error) {
	q.Normalize()
	items, total, err := s.products.List(ctx, q.Category, q.Campus, q.Keyword, q.Status, q.Page, q.PageSize)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("product list: %w", err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	return &dto.PageResult{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Remove lets the seller take down a product.
func (s *ProductService) Remove(ctx context.Context, sellerID, productID uint) (*model.Product, error) {
	p, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, util.WrapAppError(fmt.Errorf("product[id=%d] remove find: %w", productID, err), 404, constants.CodeNotFound, constants.MsgNotFound)
	}
	if p.SellerID != sellerID {
		return nil, util.NewAppError(403, constants.CodeForbidden, constants.MsgForbidden, nil)
	}
	if err := s.products.UpdateStatus(ctx, productID, constants.ProductStatusRemoved); err != nil {
		return nil, util.WrapAppError(fmt.Errorf("product[id=%d] remove: %w", productID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogProductRemoveSuccess, productID))
	p.Status = constants.ProductStatusRemoved
	return p, nil
}

// MarkSold sets the product as sold after trade completion.
func (s *ProductService) MarkSold(ctx context.Context, productID uint) error {
	if err := s.products.UpdateStatus(ctx, productID, constants.ProductStatusSold); err != nil {
		return util.WrapAppError(fmt.Errorf("product[id=%d] mark sold: %w", productID, err), 500, constants.CodeInternalError, constants.MsgInternalError)
	}
	s.logger.Info(fmt.Sprintf(constants.LogProductSoldSuccess, productID))
	return nil
}
