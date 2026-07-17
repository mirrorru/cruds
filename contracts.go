package cruds

import (
	"context"
)

// TypedTable defines the interface for typed table implementations.
// Generated typed tables must implement this interface for compile-time verification.
type TypedTable[ROW any] interface {
	Ins(ctx context.Context, tx TxProcessor, row *ROW) (*ROW, Result, error)
	Upd(ctx context.Context, tx TxProcessor, row *ROW) (*ROW, Result, error)
	One(ctx context.Context, tx TxProcessor, keys ...any) (*ROW, error)
	Del(ctx context.Context, tx TxProcessor, keys ...any) (Result, error)
	Many(ctx context.Context, tx TxProcessor, filter *Filter) ([]*ROW, error)
}

// TxProcessor описывает интерфейс выполнения SQL-запросов.
type TxProcessor interface {
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) Row
}

// Row представляет одну строку результата запроса.
type Row interface {
	Scan(dest ...any) error
}

// Result представляет результат выполнения ExecContext.
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Rows представляет курсор результатов запроса.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}
