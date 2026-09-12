package repository

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/lp/campus-market/internal/util"
)

func TestMapTradeOrderCreateError(t *testing.T) {
	duplicate := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}
	fkErr := &mysql.MySQLError{Number: 1452, Message: "foreign key"}
	plain := errors.New("boom")

	if got := mapTradeOrderCreateError(duplicate); !errors.Is(got, util.ErrConflict) {
		t.Fatalf("1062 -> %v, want ErrConflict", got)
	}
	// Non-duplicate errors must be returned untouched so callers can see them.
	if got := mapTradeOrderCreateError(fkErr); got != fkErr {
		t.Fatalf("1452 -> %v, want the original error", got)
	}
	if got := mapTradeOrderCreateError(plain); got != plain {
		t.Fatalf("plain error -> %v, want the original error", got)
	}
	if got := mapTradeOrderCreateError(nil); got != nil {
		t.Fatalf("nil -> %v, want nil", got)
	}
}

func TestEnsureTradeOrderActiveKeyDDLConstant(t *testing.T) {
	for _, want := range []string{"GENERATED ALWAYS AS", "pending", "confirmed", "uq_trade_orders_active", "CONCAT_WS"} {
		if !contains(tradeOrderKeyDDL, want) {
			t.Fatalf("DDL must contain %q: %s", want, tradeOrderKeyDDL)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
