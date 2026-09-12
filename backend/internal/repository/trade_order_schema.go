package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

// Database-level guard against duplicate active orders. A generated column
// materialises the (product_id, buyer_id) key only while the order is active
// (pending/confirmed); finished or cancelled orders produce NULL, which a
// UNIQUE index ignores, so the same buyer may order again after
// cancel/completion.
//
// The application-level "lookup then insert" check cannot be atomic, but this
// index is: at most one racing create succeeds and the others fail with MySQL
// error 1062.
const (
	tradeOrderTable     = "trade_orders"
	tradeOrderKeyColumn = "active_order_key"
	tradeOrderKeyIndex  = "uq_trade_orders_active"

	tradeOrderAddColumnDDL = "ALTER TABLE trade_orders " +
		"ADD COLUMN active_order_key VARCHAR(64) GENERATED ALWAYS AS " +
		"(CASE WHEN status IN ('pending','confirmed') " +
		"THEN CONCAT_WS(':', product_id, buyer_id) END) VIRTUAL"

	tradeOrderAddIndexDDL = "ALTER TABLE trade_orders " +
		"ADD UNIQUE KEY uq_trade_orders_active (active_order_key)"
)

// schemaStore is the subset of database operations the migration needs.
// Splitting the checks out lets the migration be exercised without a live
// database and keeps each existence probe explicit.
type schemaStore interface {
	// ColumnExists reports whether table.column is present.
	ColumnExists(table, column string) (bool, error)
	// IndexExists reports whether table has an index of the given name.
	IndexExists(table, index string) (bool, error)
	// ExecSQL runs a DDL statement.
	ExecSQL(ddl string) error
}

type gormSchemaStore struct{ db *gorm.DB }

func (g *gormSchemaStore) ColumnExists(table, column string) (bool, error) {
	var count int64
	err := g.db.Raw(
		"SELECT COUNT(1) FROM information_schema.columns "+
			"WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?",
		table, column,
	).Scan(&count).Error
	return count > 0, err
}

func (g *gormSchemaStore) IndexExists(table, index string) (bool, error) {
	var count int64
	err := g.db.Raw(
		"SELECT COUNT(1) FROM information_schema.statistics "+
			"WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		table, index,
	).Scan(&count).Error
	return count > 0, err
}

func (g *gormSchemaStore) ExecSQL(ddl string) error {
	return g.db.Exec(ddl).Error
}

// EnsureTradeOrderActiveKey makes the active-order guard real. The generated
// column and the unique index are verified and repaired independently: a
// database may have the column from an earlier partial migration without the
// index, in which case the index alone is added. It is safe to run repeatedly;
// existing order rows are not modified (the column is VIRTUAL and the index
// only reads existing values).
func EnsureTradeOrderActiveKey(db *gorm.DB) error {
	return ensureTradeOrderActiveKey(&gormSchemaStore{db: db})
}

func ensureTradeOrderActiveKey(store schemaStore) error {
	columnExists, err := store.ColumnExists(tradeOrderTable, tradeOrderKeyColumn)
	if err != nil {
		return err
	}
	if !columnExists {
		if err := store.ExecSQL(tradeOrderAddColumnDDL); err != nil {
			return err
		}
	}

	indexExists, err := store.IndexExists(tradeOrderTable, tradeOrderKeyIndex)
	if err != nil {
		return err
	}
	if !indexExists {
		// Adding the unique index fails loudly on pre-existing duplicate
		// active orders (1062) instead of silently leaving the guard half on.
		if err := store.ExecSQL(tradeOrderAddIndexDDL); err != nil {
			return err
		}
	}
	return nil
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
