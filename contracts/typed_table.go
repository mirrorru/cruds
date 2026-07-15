package contracts

import (
	"context"
	"quick-crud/filter"
)

// TypedTable defines the interface for typed table implementations.
// Generated typed tables must implement this interface for compile-time verification.
type TypedTable[ROW any] interface {
	Ins(ctx context.Context, tx TxProcessor, row *ROW) (*ROW, Result, error)
	Upd(ctx context.Context, tx TxProcessor, row *ROW) (*ROW, Result, error)
	One(ctx context.Context, tx TxProcessor, keys ...any) (*ROW, error)
	Del(ctx context.Context, tx TxProcessor, keys ...any) (Result, error)
	Many(ctx context.Context, tx TxProcessor, filter *filter.Filter) ([]*ROW, error)
}
