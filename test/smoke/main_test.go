//go:build smoke

package smoke

import (
	"database/sql"
	"os"
	"testing"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/tx_adapter"
	"github.com/mirrorru/dot"
	_ "modernc.org/sqlite"
)

var sharedDB *sql.DB

func TestMain(m *testing.M) {
	sharedDB = dot.MustMake(sql.Open("sqlite", "file::memory:?cache=shared"))
	sharedDB.SetMaxOpenConns(1)
	code := m.Run()
	sharedDB.Close()
	os.Exit(code)
}

func sharedTx() cruds.TxProcessor {
	return tx_adapter.NewDBAdapterVal(sharedDB)
}

func sharedExec(sql string) {
	dot.MustMake(sharedDB.Exec(sql))
}
