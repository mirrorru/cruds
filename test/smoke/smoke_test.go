//go:build smoke

package smoke

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/tx_adapter"

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
			db = dot.MustMake(sql.Open("sqlite", "file::memory:?cache=shared"))
			dot.MustMake(db.Exec(setupDBSQL))
			dot.MustMake(db.Exec(insertDataSQL))
		})
		return db
	}
}()

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

func (e *testEnv) TxProcessor() cruds.TxProcessor {
	return tx_adapter.NewDBAdapterVal(e.db)
}
