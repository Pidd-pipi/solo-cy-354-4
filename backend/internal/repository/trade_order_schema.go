package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

// tradeOrderActiveDDL installs the database-level guard against duplicate
// active orders. A generated column materialises the (product_id, buyer_id)
// key only while the order is active (pending/confirmed); finished or
// cancelled orders produce NULL, which a UNIQUE index ignores, so the same
// buyer may place a new order again after cancel/completion.
//
// This is what makes concurrent create requests safe: the application-level
// "lookup then insert" check cannot be atomic, but this index is, so at most
// one racing insert succeeds and the others fail with MySQL error 1062.
const (
	tradeOrderTable     = "trade_orders"
	tradeOrderKeyColumn = "active_order_key"
	tradeOrderKeyDDL    = "ALTER TABLE trade_orders " +
		"ADD COLUMN active_order_key VARCHAR(64) GENERATED ALWAYS AS " +
		"(CASE WHEN status IN ('pending','confirmed') " +
		"THEN CONCAT_WS(':', product_id, buyer_id) END) VIRTUAL, " +
		"ADD UNIQUE KEY uq_trade_orders_active (active_order_key)"
)

// EnsureTradeOrderActiveKey applies the active-order unique key once. It is
// idempotent: it only runs the ALTER when the generated column is absent.
func EnsureTradeOrderActiveKey(db *gorm.DB) error {
	var count int64
	err := db.Raw(
		"SELECT COUNT(1) FROM information_schema.columns "+
			"WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?",
		tradeOrderTable, tradeOrderKeyColumn,
	).Scan(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Exec(tradeOrderKeyDDL).Error
}

// mapTradeOrderCreateError converts a database write failure into a sentinel.
// A duplicate-key rejection (error 1062) means a concurrent request won the
// race for the one active (product, buyer) order and surfaces as ErrConflict;
// everything else is returned untouched for generic 500 handling.
func mapTradeOrderCreateError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return util.ErrConflict
	}
	return err
}
