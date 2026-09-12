package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/util"
)

// fakeOrderRepo is an in-memory OrderRepository for lifecycle tests.
type fakeOrderRepo struct {
	orders    map[uint]*model.TradeOrder
	nextID    uint
	lookupErr error // when non-nil, FindByProductBuyerStatuses fails
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: map[uint]*model.TradeOrder{}, nextID: 1}
}

func (f *fakeOrderRepo) Create(_ context.Context, o *model.TradeOrder) error {
	o.ID = f.nextID
	f.nextID++
	cp := *o
	f.orders[o.ID] = &cp
	return nil
}

func (f *fakeOrderRepo) FindByID(_ context.Context, id uint) (*model.TradeOrder, error) {
	if o, ok := f.orders[id]; ok {
		cp := *o
		return &cp, nil
	}
	return nil, util.ErrNotFound
}

func (f *fakeOrderRepo) FindByProductBuyerStatuses(_ context.Context, productID, buyerID uint, statuses []string) (*model.TradeOrder, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	for _, o := range f.orders {
		if o.ProductID != productID || o.BuyerID != buyerID {
			continue
		}
		for _, s := range statuses {
			if o.Status == s {
				cp := *o
				return &cp, nil
			}
		}
	}
	return nil, util.ErrNotFound
}

func (f *fakeOrderRepo) ListByUser(_ context.Context, userID uint, _, _ int) ([]model.TradeOrder, int64, error) {
	var out []model.TradeOrder
	for _, o := range f.orders {
		if o.BuyerID == userID || o.SellerID == userID {
			out = append(out, *o)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeOrderRepo) Transition(_ context.Context, id uint, expectStatus, toStatus string, columns map[string]interface{}) error {
	o, ok := f.orders[id]
	if !ok {
		return util.ErrNotFound
	}
	if o.Status != expectStatus {
		return util.ErrConflict
	}
	o.Status = toStatus
	for col, val := range columns {
		ts, ok := val.(time.Time)
		if !ok {
			continue
		}
		switch col {
		case "buyer_confirmed_at":
			o.BuyerConfirmedAt = &ts
		case "seller_confirmed_at":
			o.SellerConfirmedAt = &ts
		case "completed_at":
			o.CompletedAt = &ts
		}
	}
	return nil
}

func (f *fakeOrderRepo) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

func newOrderService(t *testing.T) (*TradeOrderService, *fakeOrderRepo, *fakeProductRepo) {
	t.Helper()
	orders := newFakeOrderRepo()
	products := newFakeProductRepo()
	p := &model.Product{SellerID: 200, Title: "书", Price: 10, Category: constants.ProductCategoryBooks, Condition: "全新", Campus: "东校区", TradeLocation: "东门", Status: constants.ProductStatusOnSale}
	if err := products.Create(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	return NewTradeOrderService(orders, products, slog.Default()), orders, products
}

func appStatus(t *testing.T, err error) int {
	t.Helper()
	var ae *util.AppError
	if errors.As(err, &ae) {
		return ae.Status
	}
	t.Fatalf("expected AppError, got %v", err)
	return 0
}

// Path 1: create — a new order starts in the initial (pending) status.
func TestTradeOrderCreateLifecycle(t *testing.T) {
	svc, orders, _ := newOrderService(t)
	buyer := &model.User{ID: 100}

	order, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.Status != constants.TradeStatusPending {
		t.Fatalf("new order status = %q, want pending", order.Status)
	}
	if len(orders.orders) != 1 {
		t.Fatalf("normal create produced %d stored orders, want exactly 1", len(orders.orders))
	}
	stored := orders.orders[order.ID]
	if stored.BuyerConfirmedAt != nil || stored.SellerConfirmedAt != nil || stored.CompletedAt != nil {
		t.Fatalf("new order must have no confirmation timestamps")
	}

	// Duplicate active order is rejected and writes nothing.
	before := len(orders.orders)
	if _, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1}); appStatus(t, err) != 409 {
		t.Fatalf("duplicate order: got status %d, want 409", appStatus(t, err))
	}
	if len(orders.orders) != before {
		t.Fatalf("duplicate create changed order count: %d -> %d", before, len(orders.orders))
	}
	// Buying own product is rejected before any lookup/write.
	if _, err := svc.Create(context.Background(), &model.User{ID: 200}, &dto.CreateTradeOrderRequest{ProductID: 1}); appStatus(t, err) != 400 {
		t.Fatalf("own product: got %d, want 400", appStatus(t, err))
	}
	// Missing product is a 404.
	if _, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 999}); appStatus(t, err) != 404 {
		t.Fatalf("missing product: got %d, want 404", appStatus(t, err))
	}
	if len(orders.orders) != 1 {
		t.Fatalf("failed branches created orders: total = %d, want 1", len(orders.orders))
	}
}

// A lookup failure during the duplicate check must abort creation with a 500
// and never insert an order.
func TestTradeOrderCreateDuplicateLookupFailure(t *testing.T) {
	svc, orders, _ := newOrderService(t)
	orders.lookupErr = errors.New("db connection lost")

	before := len(orders.orders)
	_, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if appStatus(t, err) != 500 {
		t.Fatalf("lookup failure: got status %d, want 500", appStatus(t, err))
	}
	if !strings.Contains(err.Error(), "product=1") || !strings.Contains(err.Error(), "buyer=100") {
		t.Fatalf("error must identify product and buyer for triage, got: %v", err)
	}
	if !strings.Contains(err.Error(), "db connection lost") {
		t.Fatalf("error must preserve the underlying cause, got: %v", err)
	}
	if len(orders.orders) != before {
		t.Fatalf("create after lookup failure wrote %d orders, want 0 extra", len(orders.orders)-before)
	}

	// Once the lookup recovers, normal creation succeeds and writes exactly one.
	orders.lookupErr = nil
	if _, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1}); err != nil {
		t.Fatalf("create after recovery: %v", err)
	}
	if len(orders.orders) != 1 {
		t.Fatalf("recovered create total orders = %d, want exactly 1", len(orders.orders))
	}
}

// Paths 1-2-3 together: pending --buyer--> confirmed --seller--> completed,
// timestamps stamped and the product taken off sale atomically.
func TestTradeOrderConfirmHappyPath(t *testing.T) {
	svc, orders, products := newOrderService(t)
	order, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatal(err)
	}

	confirmed, err := svc.BuyerConfirm(context.Background(), 100, order.ID)
	if err != nil {
		t.Fatalf("buyer confirm: %v", err)
	}
	if confirmed.Status != constants.TradeStatusConfirmed {
		t.Fatalf("after buyer confirm status = %q, want confirmed", confirmed.Status)
	}
	if orders.orders[order.ID].BuyerConfirmedAt == nil {
		t.Fatalf("buyer_confirmed_at must be set")
	}
	if orders.orders[order.ID].SellerConfirmedAt != nil {
		t.Fatalf("seller_confirmed_at must stay unset after buyer confirm")
	}

	completed, err := svc.SellerConfirm(context.Background(), 200, order.ID)
	if err != nil {
		t.Fatalf("seller confirm: %v", err)
	}
	if completed.Status != constants.TradeStatusCompleted {
		t.Fatalf("after seller confirm status = %q, want completed", completed.Status)
	}
	stored := orders.orders[order.ID]
	if stored.SellerConfirmedAt == nil || stored.CompletedAt == nil {
		t.Fatalf("seller_confirmed_at and completed_at must be set: %+v", stored)
	}
	if products.products[1].Status != constants.ProductStatusSold {
		t.Fatalf("product status = %q, want sold", products.products[1].Status)
	}
}

// Path 2 rejections: only the buyer, only from pending.
func TestTradeOrderBuyerConfirmGuards(t *testing.T) {
	svc, orders, _ := newOrderService(t)
	order, _ := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})

	// Seller cannot confirm as buyer -> 403 (role checked before status).
	if _, err := svc.BuyerConfirm(context.Background(), 200, order.ID); appStatus(t, err) != 403 {
		t.Fatalf("seller buyer-confirm: got %d, want 403", appStatus(t, err))
	}
	// Stranger -> 403.
	if _, err := svc.BuyerConfirm(context.Background(), 999, order.ID); appStatus(t, err) != 403 {
		t.Fatalf("stranger buyer-confirm: got %d, want 403", appStatus(t, err))
	}
	if _, err := svc.BuyerConfirm(context.Background(), 100, order.ID); err != nil {
		t.Fatal(err)
	}
	// Repeat confirmation after status moved -> 409.
	if _, err := svc.BuyerConfirm(context.Background(), 100, order.ID); appStatus(t, err) != 409 {
		t.Fatalf("repeat buyer-confirm: got %d, want 409", appStatus(t, err))
	}
	// Missing order -> 404.
	if _, err := svc.BuyerConfirm(context.Background(), 100, 999); appStatus(t, err) != 404 {
		t.Fatalf("missing order: got %d, want 404", appStatus(t, err))
	}
	_ = orders
}

// Path 3 rejections: only the seller, only from confirmed.
func TestTradeOrderSellerConfirmGuards(t *testing.T) {
	svc, _, _ := newOrderService(t)
	order, _ := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})

	// From pending the seller-confirm transition is illegal -> 409.
	if _, err := svc.SellerConfirm(context.Background(), 200, order.ID); appStatus(t, err) != 409 {
		t.Fatalf("seller confirm from pending: got %d, want 409", appStatus(t, err))
	}
	// Buyer cannot seller-confirm (even on a confirmed order) -> 403.
	if _, err := svc.BuyerConfirm(context.Background(), 100, order.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SellerConfirm(context.Background(), 100, order.ID); appStatus(t, err) != 403 {
		t.Fatalf("buyer seller-confirm: got %d, want 403", appStatus(t, err))
	}
	// Proper completion.
	if _, err := svc.SellerConfirm(context.Background(), 200, order.ID); err != nil {
		t.Fatalf("seller confirm: %v", err)
	}
	// Completing twice -> 409.
	if _, err := svc.SellerConfirm(context.Background(), 200, order.ID); appStatus(t, err) != 409 {
		t.Fatalf("repeat seller-confirm: got %d, want 409", appStatus(t, err))
	}
}

// Path 4: cancel — pending order cancellable by either party, never afterwards.
func TestTradeOrderCancelLifecycle(t *testing.T) {
	// Buyer cancels.
	svc, orders, _ := newOrderService(t)
	order, _ := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	cancelled, err := svc.Cancel(context.Background(), 100, order.ID)
	if err != nil {
		t.Fatalf("buyer cancel: %v", err)
	}
	if cancelled.Status != constants.TradeStatusCancelled {
		t.Fatalf("status = %q, want cancelled", cancelled.Status)
	}
	if orders.orders[order.ID].CompletedAt != nil {
		t.Fatalf("cancel must not stamp completed_at")
	}
	// Cancelling twice -> 409.
	if _, err := svc.Cancel(context.Background(), 100, order.ID); appStatus(t, err) != 409 {
		t.Fatalf("repeat cancel: got %d, want 409", appStatus(t, err))
	}

	// Seller may cancel another pending order.
	svc2, _, _ := newOrderService(t)
	order2, _ := svc2.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if _, err := svc2.Cancel(context.Background(), 200, order2.ID); err != nil {
		t.Fatalf("seller cancel: %v", err)
	}

	// A confirmed order can no longer be cancelled.
	svc3, _, _ := newOrderService(t)
	order3, _ := svc3.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if _, err := svc3.BuyerConfirm(context.Background(), 100, order3.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc3.Cancel(context.Background(), 100, order3.ID); appStatus(t, err) != 409 {
		t.Fatalf("cancel confirmed: got %d, want 409", appStatus(t, err))
	}
	// Stranger cannot cancel a pending order -> 403.
	svc4, _, _ := newOrderService(t)
	order4, _ := svc4.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if _, err := svc4.Cancel(context.Background(), 999, order4.ID); appStatus(t, err) != 403 {
		t.Fatalf("stranger cancel: got %d, want 403", appStatus(t, err))
	}
}
