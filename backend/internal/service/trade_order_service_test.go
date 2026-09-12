package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/dto"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/util"
)

// fakeOrderRepo is an in-memory OrderRepository for lifecycle tests. The
// mutex makes Transition an atomic compare-and-set (mirroring the SQL
// WHERE status = ? guard) so concurrent calls can race deterministically,
// and Transaction snapshots/restores state like a real DB rollback.
type fakeOrderRepo struct {
	mu        sync.Mutex
	orders    map[uint]*model.TradeOrder
	nextID    uint
	lookupErr error // when non-nil, FindByProductBuyerStatuses fails
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: map[uint]*model.TradeOrder{}, nextID: 1}
}

func (f *fakeOrderRepo) snapshot() map[uint]*model.TradeOrder {
	cp := make(map[uint]*model.TradeOrder, len(f.orders))
	for id, o := range f.orders {
		row := *o
		cp[id] = &row
	}
	return cp
}

func (f *fakeOrderRepo) Create(_ context.Context, o *model.TradeOrder) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Mirror the database uq_trade_orders_active key: one active (pending or
	// confirmed) order per (product, buyer). Finished/cancelled rows do not
	// collide. Holding the lock across check+insert makes it a real race test.
	for _, e := range f.orders {
		if e.ProductID == o.ProductID && e.BuyerID == o.BuyerID &&
			(e.Status == constants.TradeStatusPending || e.Status == constants.TradeStatusConfirmed) {
			return util.ErrConflict
		}
	}
	o.ID = f.nextID
	f.nextID++
	cp := *o
	f.orders[o.ID] = &cp
	return nil
}

func (f *fakeOrderRepo) FindByID(_ context.Context, id uint) (*model.TradeOrder, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if o, ok := f.orders[id]; ok {
		cp := *o
		return &cp, nil
	}
	return nil, util.ErrNotFound
}

func (f *fakeOrderRepo) FindByProductBuyerStatuses(_ context.Context, productID, buyerID uint, statuses []string) (*model.TradeOrder, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
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
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []model.TradeOrder
	for _, o := range f.orders {
		if o.BuyerID == userID || o.SellerID == userID {
			out = append(out, *o)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeOrderRepo) Transition(_ context.Context, id uint, expectStatus, toStatus string, columns map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()
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

// Transaction mirrors a database transaction: mutations commit on success and
// are rolled back to the entry snapshot on error.
func (f *fakeOrderRepo) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	f.mu.Lock()
	snapshot := f.snapshot()
	f.mu.Unlock()
	if err := fn(ctx); err != nil {
		f.mu.Lock()
		f.orders = snapshot
		f.mu.Unlock()
		return err
	}
	return nil
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

// TestCreateDuplicateLookupFailureNoOrder is a standalone, repeatable test
// for the triage rule: when the duplicate-order lookup itself fails, creation
// must abort with a 500 and must not insert an order.
func TestCreateDuplicateLookupFailureNoOrder(t *testing.T) {
	svc, orders, _ := newOrderService(t)
	dbErr := errors.New("lookup: connection reset")
	orders.lookupErr = dbErr

	const tries = 5
	for i := 0; i < tries; i++ {
		before := len(orders.orders)
		_, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
		if appStatus(t, err) != 500 {
			t.Fatalf("attempt %d: got status %d, want 500 (err=%v)", i+1, appStatus(t, err), err)
		}
		if !strings.Contains(err.Error(), "product=1") || !strings.Contains(err.Error(), "buyer=100") {
			t.Fatalf("attempt %d: error must carry product and buyer ids: %v", i+1, err)
		}
		if !errors.Is(err, dbErr) {
			t.Fatalf("attempt %d: underlying lookup error must be preserved: %v", i+1, err)
		}
		if got := len(orders.orders); got != before {
			t.Fatalf("attempt %d: order count changed %d -> %d", i+1, before, got)
		}
	}
	if got := len(orders.orders); got != 0 {
		t.Fatalf("failed lookups created %d orders, want 0", got)
	}

	// The product stays on sale (nothing was reserved) and a healthy lookup
	// afterwards creates exactly one order.
	orders.lookupErr = nil
	order, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatalf("create after recovery failed: %v", err)
	}
	if order.Status != constants.TradeStatusPending || len(orders.orders) != 1 {
		t.Fatalf("post-recovery state wrong: status=%q orders=%d", order.Status, len(orders.orders))
	}
}

// TestBuyerConfirmConcurrentOnlyOneSucceeds is a standalone, repeatable test:
// the same buyer confirming one pending order concurrently any number of
// times must yield exactly one success; every other call is rejected as 409
// and the order ends in confirmed with exactly one confirmation timestamp.
func TestBuyerConfirmConcurrentOnlyOneSucceeds(t *testing.T) {
	const goroutines = 16
	for iteration := 0; iteration < 20; iteration++ {
		svc, orders, _ := newOrderService(t)
		order, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
		if err != nil {
			t.Fatalf("iteration %d: create: %v", iteration, err)
		}

		start := make(chan struct{})
		var wg sync.WaitGroup
		var okMu sync.Mutex
		successes, conflicts := 0, 0
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				<-start
				_, err := svc.BuyerConfirm(context.Background(), 100, order.ID)
				okMu.Lock()
				defer okMu.Unlock()
				switch {
				case err == nil:
					successes++
				case appStatus(t, err) == 409:
					conflicts++
				default:
					t.Errorf("iteration %d: unexpected confirm error: %v", iteration, err)
				}
			}()
		}
		close(start)
		wg.Wait()

		if successes != 1 {
			t.Fatalf("iteration %d: successes=%d, want exactly 1 (conflicts=%d)", iteration, successes, conflicts)
		}
		if conflicts != goroutines-1 {
			t.Fatalf("iteration %d: conflicts=%d, want %d", iteration, conflicts, goroutines-1)
		}
		stored := orders.orders[order.ID]
		if stored.Status != constants.TradeStatusConfirmed || stored.BuyerConfirmedAt == nil {
			t.Fatalf("iteration %d: stored order wrong after race: %+v", iteration, stored)
		}
		if stored.SellerConfirmedAt != nil || stored.CompletedAt != nil {
			t.Fatalf("iteration %d: buyer confirm stamped unrelated timestamps: %+v", iteration, stored)
		}
	}
}

// TestSellerConfirmProductWriteFailureRollsBack is a standalone, repeatable
// test: if writing the product status fails inside the seller-confirm
// transaction, the order must stay confirmed (not completed), timestamps
// must not be persisted and the caller receives a 500. Once the write heals,
// the same order can complete normally exactly once.
func TestSellerConfirmProductWriteFailureRollsBack(t *testing.T) {
	svc, orders, products := newOrderService(t)
	order, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BuyerConfirm(context.Background(), 100, order.ID); err != nil {
		t.Fatalf("buyer confirm: %v", err)
	}
	stored := orders.orders[order.ID]
	if stored.Status != constants.TradeStatusConfirmed {
		t.Fatalf("precondition: status=%q, want confirmed", stored.Status)
	}

	writeErr := errors.New("product update: disk full")
	products.updateStatusErr = writeErr
	_, err = svc.SellerConfirm(context.Background(), 200, order.ID)
	if appStatus(t, err) != 500 {
		t.Fatalf("got status %d, want 500 (err=%v)", appStatus(t, err), err)
	}
	if !errors.Is(err, writeErr) {
		t.Fatalf("underlying product write error must be preserved: %v", err)
	}

	stored = orders.orders[order.ID]
	if stored.Status != constants.TradeStatusConfirmed {
		t.Fatalf("order status after failed product write = %q, want confirmed (rollback)", stored.Status)
	}
	if stored.SellerConfirmedAt != nil || stored.CompletedAt != nil {
		t.Fatalf("rollback must clear confirmation timestamps: %+v", stored)
	}
	if products.products[1].Status != constants.ProductStatusOnSale {
		t.Fatalf("product status = %q, want on_sale", products.products[1].Status)
	}

	// Healed write: the still-confirmed order completes once, atomically.
	products.updateStatusErr = nil
	completed, err := svc.SellerConfirm(context.Background(), 200, order.ID)
	if err != nil {
		t.Fatalf("seller confirm after recovery: %v", err)
	}
	if completed.Status != constants.TradeStatusCompleted {
		t.Fatalf("status = %q, want completed", completed.Status)
	}
	stored = orders.orders[order.ID]
	if stored.SellerConfirmedAt == nil || stored.CompletedAt == nil {
		t.Fatalf("completion must stamp seller/completed times: %+v", stored)
	}
	if products.products[1].Status != constants.ProductStatusSold {
		t.Fatalf("product status = %q, want sold", products.products[1].Status)
	}

	// The transition cannot be applied twice.
	if _, err := svc.SellerConfirm(context.Background(), 200, order.ID); appStatus(t, err) != 409 {
		t.Fatalf("repeat seller confirm: got %d, want 409", appStatus(t, err))
	}
}

// TestCreateConcurrentOnlyOneSucceeds verifies the race fix: many concurrent
// create requests for the same product and buyer must result in exactly one
// pending order; every losing request is a 409 conflict with no extra rows.
func TestCreateConcurrentOnlyOneSucceeds(t *testing.T) {
	const goroutines = 16
	for iteration := 0; iteration < 20; iteration++ {
		svc, orders, _ := newOrderService(t)
		start := make(chan struct{})
		var wg sync.WaitGroup
		var statMu sync.Mutex
		successes, conflicts, others := 0, 0, 0
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				<-start
				_, err := svc.Create(context.Background(), &model.User{ID: 100}, &dto.CreateTradeOrderRequest{ProductID: 1})
				statMu.Lock()
				defer statMu.Unlock()
				switch {
				case err == nil:
					successes++
				case appStatus(t, err) == 409:
					conflicts++
				default:
					others++
					t.Errorf("iteration %d: unexpected create error: %v", iteration, err)
				}
			}()
		}
		close(start)
		wg.Wait()

		if successes != 1 || conflicts != goroutines-1 || others != 0 {
			t.Fatalf("iteration %d: successes=%d conflicts=%d others=%d, want 1/%d/0",
				iteration, successes, conflicts, others, goroutines-1)
		}
		if len(orders.orders) != 1 {
			t.Fatalf("iteration %d: stored orders = %d, want exactly 1", iteration, len(orders.orders))
		}
		var pendingCount int
		for _, o := range orders.orders {
			if o.Status == constants.TradeStatusPending {
				pendingCount++
			}
		}
		if pendingCount != 1 {
			t.Fatalf("iteration %d: pending orders = %d, want 1", iteration, pendingCount)
		}
	}
}

// TestCreateAgainAfterCancel preserves the existing rule: once the active
// order is cancelled the (product, buyer) key frees up and a new order may be
// placed; the second order is a distinct pending row.
func TestCreateAgainAfterCancel(t *testing.T) {
	svc, orders, _ := newOrderService(t)
	buyer := &model.User{ID: 100}

	first, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := svc.Cancel(context.Background(), 100, first.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	second, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatalf("re-order after cancel: %v", err)
	}
	if second.ID == first.ID || second.Status != constants.TradeStatusPending {
		t.Fatalf("re-order must be a new pending order: first=%d second=%+v", first.ID, second)
	}
	if len(orders.orders) != 2 {
		t.Fatalf("stored orders = %d, want 2 (one cancelled, one pending)", len(orders.orders))
	}

	// While the new order stays active another duplicate is still rejected.
	if _, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1}); appStatus(t, err) != 409 {
		t.Fatalf("duplicate on re-ordered product: got %d, want 409", appStatus(t, err))
	}
}

// TestCreateAgainAfterCompletion keeps post-completion behaviour unchanged:
// the active-order key frees when the order completes, and rejection then
// comes from the existing on-sale product rule. A fresh on-sale product is
// still orderable by the same buyer.
func TestCreateAgainAfterCompletion(t *testing.T) {
	svc, orders, products := newOrderService(t)
	buyer := &model.User{ID: 100}
	order, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BuyerConfirm(context.Background(), 100, order.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SellerConfirm(context.Background(), 200, order.ID); err != nil {
		t.Fatal(err)
	}

	// Product 1 is sold now: the completed order no longer occupies the key,
	// but the unchanged on-sale rule rejects the request with 409.
	if _, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: 1}); appStatus(t, err) != 409 {
		t.Fatalf("re-order sold product: got %d, want 409", appStatus(t, err))
	}
	if len(orders.orders) != 1 {
		t.Fatalf("re-order attempt after completion wrote %d orders, want 1", len(orders.orders))
	}

	// A different product being sold by the same seller is independently
	// orderable by the same buyer.
	other := &model.Product{SellerID: 200, Title: "另一本书", Price: 5, Category: constants.ProductCategoryBooks, Condition: "全新", Campus: "东校区", TradeLocation: "东门", Status: constants.ProductStatusOnSale}
	if err := products.Create(context.Background(), other); err != nil {
		t.Fatal(err)
	}
	again, err := svc.Create(context.Background(), buyer, &dto.CreateTradeOrderRequest{ProductID: other.ID})
	if err != nil {
		t.Fatalf("order a different on-sale product: %v", err)
	}
	if again.Status != constants.TradeStatusPending {
		t.Fatalf("new order status = %q, want pending", again.Status)
	}
}
