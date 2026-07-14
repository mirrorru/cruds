package contracts

import "context"

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
