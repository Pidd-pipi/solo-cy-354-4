package tradestate

import (
	"errors"
	"testing"
	"time"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/model"
)

func makeOrder(status string) *model.TradeOrder {
	return &model.TradeOrder{ID: 1, ProductID: 10, BuyerID: 100, SellerID: 200, Status: status}
}

func TestGuardHappyPaths(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		action      string
		userID      uint
		wantTo      string
		wantProduct string
	}{
		{name: "buyer confirms pending", status: constants.TradeStatusPending, action: ActionBuyerConfirm, userID: 100, wantTo: constants.TradeStatusConfirmed},
		{name: "seller confirms confirmed", status: constants.TradeStatusConfirmed, action: ActionSellerConfirm, userID: 200, wantTo: constants.TradeStatusCompleted, wantProduct: constants.ProductStatusSold},
		{name: "buyer cancels pending", status: constants.TradeStatusPending, action: ActionCancel, userID: 100, wantTo: constants.TradeStatusCancelled},
		{name: "seller cancels pending", status: constants.TradeStatusPending, action: ActionCancel, userID: 200, wantTo: constants.TradeStatusCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := Guard(makeOrder(tt.status), tt.action, tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rule.From != tt.status || rule.To != tt.wantTo {
				t.Fatalf("got %s->%s, want %s->%s", rule.From, rule.To, tt.status, tt.wantTo)
			}
			if rule.ProductOnComplete != tt.wantProduct {
				t.Fatalf("product side effect = %q, want %q", rule.ProductOnComplete, tt.wantProduct)
			}
		})
	}
}

func TestGuardRejections(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		action  string
		userID  uint
		wantErr error
	}{
		{name: "seller cannot buyer-confirm", status: constants.TradeStatusPending, action: ActionBuyerConfirm, userID: 200, wantErr: ErrIllegalRole},
		{name: "buyer cannot seller-confirm", status: constants.TradeStatusConfirmed, action: ActionSellerConfirm, userID: 100, wantErr: ErrIllegalRole},
		{name: "stranger cannot cancel", status: constants.TradeStatusPending, action: ActionCancel, userID: 999, wantErr: ErrIllegalRole},
		{name: "buyer confirm requires pending", status: constants.TradeStatusConfirmed, action: ActionBuyerConfirm, userID: 100, wantErr: ErrIllegalTransition},
		{name: "buyer confirm on completed", status: constants.TradeStatusCompleted, action: ActionBuyerConfirm, userID: 100, wantErr: ErrIllegalTransition},
		{name: "seller confirm requires confirmed", status: constants.TradeStatusPending, action: ActionSellerConfirm, userID: 200, wantErr: ErrIllegalTransition},
		{name: "seller confirm on cancelled", status: constants.TradeStatusCancelled, action: ActionSellerConfirm, userID: 200, wantErr: ErrIllegalTransition},
		{name: "cancel only from pending (confirmed)", status: constants.TradeStatusConfirmed, action: ActionCancel, userID: 100, wantErr: ErrIllegalTransition},
		{name: "cancel completed", status: constants.TradeStatusCompleted, action: ActionCancel, userID: 100, wantErr: ErrIllegalTransition},
		{name: "cancel cancelled", status: constants.TradeStatusCancelled, action: ActionCancel, userID: 100, wantErr: ErrIllegalTransition},
		{name: "unknown action", status: constants.TradeStatusPending, action: "explode", userID: 100, wantErr: ErrIllegalTransition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Guard(makeOrder(tt.status), tt.action, tt.userID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleColumnUpdates(t *testing.T) {
	now := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
	buyerRule, err := Guard(makeOrder(constants.TradeStatusPending), ActionBuyerConfirm, 100)
	if err != nil {
		t.Fatal(err)
	}
	updates := buyerRule.ColumnUpdates(now)
	if _, ok := updates["buyer_confirmed_at"]; !ok {
		t.Fatalf("buyer transition must stamp buyer_confirmed_at: %v", updates)
	}
	if _, ok := updates["seller_confirmed_at"]; ok {
		t.Fatalf("buyer transition must not touch seller_confirmed_at: %v", updates)
	}
	if _, ok := updates["completed_at"]; ok {
		t.Fatalf("buyer transition must not touch completed_at: %v", updates)
	}

	sellerRule, err := Guard(makeOrder(constants.TradeStatusConfirmed), ActionSellerConfirm, 200)
	if err != nil {
		t.Fatal(err)
	}
	updates = sellerRule.ColumnUpdates(now)
	for _, col := range []string{"seller_confirmed_at", "completed_at"} {
		if _, ok := updates[col]; !ok {
			t.Fatalf("seller transition must stamp %s: %v", col, updates)
		}
	}
	if _, ok := updates["buyer_confirmed_at"]; ok {
		t.Fatalf("seller transition must not touch buyer_confirmed_at: %v", updates)
	}

	cancelRule, err := Guard(makeOrder(constants.TradeStatusPending), ActionCancel, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(cancelRule.ColumnUpdates(now)) != 0 {
		t.Fatalf("cancel must not stamp any timestamp: %v", updates)
	}
}

func TestInitialAndActiveStatuses(t *testing.T) {
	if Initial() != constants.TradeStatusPending {
		t.Fatalf("new orders must start pending, got %q", Initial())
	}
	got := ActiveStatuses()
	want := []string{constants.TradeStatusPending, constants.TradeStatusConfirmed}
	if len(got) != len(want) {
		t.Fatalf("active statuses = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("active statuses = %v, want %v", got, want)
		}
	}
}
