//go:build smoke

package smoke

import (
	"context"
	"database/sql"
	"log/slog"
	quick_crud "quick-crud"
	"quick-crud/tx_adapter"
	"sync"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/mirrorru/dot"
	_ "modernc.org/sqlite"
)

var sqliteConn func() *sql.DB = func() func() *sql.DB {
	var (
		once sync.Once
		db   *sql.DB
	)
	return func() *sql.DB {
		once.Do(func() {
			db = dot.MustMake(sql.Open("sqlite", ":memory:"))
			dot.MustMake(db.Exec(setupDBSQL))
			dot.MustMake(db.Exec(insertDataSQL))
		})
		return db
	}
}()

func TestMain(m *testing.M) {
	slog.Info("DB created")
	m.Run()
	slog.Info("TESTING DONE")
}

const setupDBSQL = `
	create table id_name_age_filled (
	    id integer primary key autoincrement not null,
	    name text not null,
	    age int null
	);

	create table id_name_age_row (
	    id integer primary key autoincrement not null,
	    name text not null,
	    age int null
	);

`
const insertDataSQL = `
	insert into id_name_age_filled (name, age) values 
		('Unaged', null),
		('Alice', 11),
		('Bob', 22);
`

type testEnv struct {
	ctx context.Context
	db  *sql.DB
}

func newTestEnv(t *testing.T) *testEnv {
	_ = minimock.NewController(t)

	return &testEnv{
		ctx: context.Background(),
		db:  sqliteConn(),
	}
}

func (e *testEnv) TxProcessor() quick_crud.TxProcessor {
	return tx_adapter.NewDBAdapterVal(e.db)
}
