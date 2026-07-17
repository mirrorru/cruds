package tx_adapter

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	qc "github.com/mirrorru/cruds"
)

// PGXPoolAdapter адаптирует *pgxpool.Pool к интерфейсу TxProcessor.
// EN: PGXPoolAdapter adapts *pgxpool.Pool to the TxProcessor interface.
type PGXPoolAdapter struct {
	pool *pgxpool.Pool
}

var _ qc.TxProcessor = PGXPoolAdapter{}

// NewPGXPoolAdapterVal создаёт адаптер для *pgxpool.Pool к интерфейсу TxProcessor.
// EN: NewPGXPoolAdapterVal creates an adapter from *pgxpool.Pool to the TxProcessor interface.
func NewPGXPoolAdapterVal(pool *pgxpool.Pool) PGXPoolAdapter {
	return PGXPoolAdapter{pool: pool}
}

// ExecContext выполняет запрос, не возвращающий строк.
// EN: ExecContext executes a query that does not return rows.
func (a PGXPoolAdapter) ExecContext(ctx context.Context, query string, args ...any) (qc.Result, error) {
	tag, err := a.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxResult{tag: tag}, nil
}

// QueryContext выполняет запрос, возвращающий строки.
// EN: QueryContext executes a query that returns rows.
func (a PGXPoolAdapter) QueryContext(ctx context.Context, query string, args ...any) (qc.Rows, error) {
	rows, err := a.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PgxRows{rows: rows}, nil
}

// QueryRowContext выполняет запрос, возвращающий одну строку.
// EN: QueryRowContext executes a query that returns a single row.
func (a PGXPoolAdapter) QueryRowContext(ctx context.Context, query string, args ...any) qc.Row {
	return a.pool.QueryRow(ctx, query, args...)
}
