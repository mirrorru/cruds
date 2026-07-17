//go:build functional

package functional

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/tx_adapter"
)

var sharedPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	dsn := os.Getenv("TEST_PG_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "TEST_PG_DSN must be set")
		os.Exit(1)
	}
	var err error
	sharedPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgxpool.New: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	sharedPool.Close()
	os.Exit(code)
}

func sharedTx() cruds.TxProcessor {
	return tx_adapter.NewPGXPoolAdapterVal(sharedPool)
}

func sharedExec(sql string) {
	_, err := sharedPool.Exec(context.Background(), sql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exec error: %v\n", err)
		os.Exit(1)
	}
}
