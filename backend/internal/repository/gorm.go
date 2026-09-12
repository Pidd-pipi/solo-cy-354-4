// Package repository implements the GORM repository layer for campus-market.
package repository

import (
	"context"
	"errors"

	"github.com/lp/campus-market/internal/util"
	"gorm.io/gorm"
)

// txKey is the context key under which an in-flight transaction lives.
type txKey struct{}

// WithTx attaches a GORM transaction to the context so repository methods
// execute within it when present.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// db returns the transaction-bound database handle if one exists in the
// context, otherwise it returns the repository's base handle.
func db(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}

// Transaction runs fn inside a GORM transaction. The transaction handle is
// injected into txCtx so all repository writes participate atomically.
func Transaction(ctx context.Context, database *gorm.DB, fn func(txCtx context.Context) error) error {
	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}

// normalizeError converts GORM errors into sentinel repository errors.
func normalizeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.ErrNotFound
	}
	return err
}
