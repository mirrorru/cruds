//go:build smoke

package repo

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/mirrorru/crudquick/dialect"
	"github.com/mirrorru/crudquick/test/gen/model"
	"github.com/mirrorru/crudquick/tx_adapter"

	qc "github.com/mirrorru/crudquick"

	"github.com/mirrorru/dot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	create table user_row (
	    id integer primary key autoincrement not null,
	    name text not null,
	    age int not null
	);

	create table products (
	    id integer primary key autoincrement not null,
	    name text not null,
	    price real not null,
	    stock int not null
	);
`

const insertDataSQL = `
	insert into user_row (name, age) values 
		('Alice', 30),
		('Bob', 25),
		('Charlie', 35);

	insert into products (name, price, stock) values 
		('Widget', 19.99, 100),
		('Gadget', 49.99, 50),
		('Thingamajig', 29.99, 75);
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

func (e *testEnv) TxProcessor() qc.TxProcessor {
	return tx_adapter.NewDBAdapterVal(e.db)
}

func TestTableUserRow_Interface(t *testing.T) {
	t.Parallel()
	var _ qc.TypedTable[model.UserRow] = (*TableUserRow)(nil)
}

func TestTableProductRow_Interface(t *testing.T) {
	t.Parallel()
	var _ qc.TypedTable[model.ProductRow] = (*TableProductRow)(nil)
}

func TestTableUserRow_Internals(t *testing.T) {
	t.Parallel()
	table := NewTableUserRow(dialect.SQLiteDialect{})
	internals := table.Internals()

	assert.Equal(t, "user_row", internals.TableInfo.SQLName)
	require.Len(t, internals.TableInfo.Fields, 3)
	assert.Equal(t, "id", internals.TableInfo.Fields[0].SQLName)
	assert.Equal(t, "name", internals.TableInfo.Fields[1].SQLName)
	assert.Equal(t, "age", internals.TableInfo.Fields[2].SQLName)
}

func TestTableProductRow_Internals(t *testing.T) {
	t.Parallel()
	table := NewTableProductRow(dialect.SQLiteDialect{})
	internals := table.Internals()

	assert.Equal(t, "products", internals.TableInfo.SQLName)
	require.Len(t, internals.TableInfo.Fields, 4)
	assert.Equal(t, "id", internals.TableInfo.Fields[0].SQLName)
	assert.Equal(t, "name", internals.TableInfo.Fields[1].SQLName)
	assert.Equal(t, "price", internals.TableInfo.Fields[2].SQLName)
	assert.Equal(t, "stock", internals.TableInfo.Fields[3].SQLName)
}

func TestTableUserRow_One(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := NewTableUserRow(dialect.SQLiteDialect{})
	tx := env.TxProcessor()

	row, err := table.One(env.ctx, tx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, row.ID)
	assert.Equal(t, "Alice", row.Name)
	assert.Equal(t, 30, row.Age)
}

func TestTableProductRow_One(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := NewTableProductRow(dialect.SQLiteDialect{})
	tx := env.TxProcessor()

	row, err := table.One(env.ctx, tx, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, row.ID)
	assert.Equal(t, "Widget", row.Name)
	assert.Equal(t, 19.99, row.Price)
	assert.Equal(t, 100, row.Stock)
}

func TestTableUserRow_Many(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := NewTableUserRow(dialect.SQLiteDialect{})
	tx := env.TxProcessor()

	rows, err := table.Many(env.ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 3)

	rows, err = table.Many(env.ctx, tx, &qc.Filter{
		Offset: 1,
		Limit:  2,
	})
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	rows, err = table.Many(env.ctx, tx, &qc.Filter{
		Range: qc.ConditionNode{
			FieldIdx: 0,
			Op:       qc.CmdEq,
			Value:    2,
		},
	})
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "Bob", rows[0].Name)
}

func TestTableProductRow_Many(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := NewTableProductRow(dialect.SQLiteDialect{})
	tx := env.TxProcessor()

	rows, err := table.Many(env.ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 3)

	rows, err = table.Many(env.ctx, tx, &qc.Filter{
		Range: qc.ConditionNode{
			FieldIdx: 2,
			Op:       qc.CmdGt,
			Value:    25.0,
		},
	})
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestTableUserRow_AllCRUD(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := NewTableUserRow(dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	ctx := env.ctx

	alice, _, err := table.Ins(ctx, tx, &model.UserRow{Name: "Alice", Age: 1})
	require.NoError(t, err)
	require.Equal(t, "Alice", alice.Name)
	require.Equal(t, 1, alice.Age)
	require.Greater(t, alice.ID, 0)

	alice2, err := table.One(ctx, tx, alice.ID)
	assert.NotSame(t, alice, alice2)
	assert.Equal(t, *alice, *alice2)

	alice2.Age = 11
	alice, _, err = table.Upd(ctx, tx, alice2)
	assert.NotSame(t, alice2, alice)
	assert.Equal(t, *alice2, *alice)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 1)

	found := false
	for _, r := range rows {
		if r.ID == alice.ID {
			found = true
			assert.Equal(t, 11, r.Age)
		}
	}
	require.True(t, found, "Alice should be in the list")

	_, err = table.Del(ctx, tx, alice.ID)
	require.NoError(t, err)

	_, err = table.One(ctx, tx, alice.ID)
	require.Error(t, err)
}

func TestTableProductRow_AllCRUD(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := NewTableProductRow(dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	ctx := env.ctx

	widget, _, err := table.Ins(ctx, tx, &model.ProductRow{Name: "Widget", Price: 19.99, Stock: 100})
	require.NoError(t, err)
	require.Equal(t, "Widget", widget.Name)
	require.Equal(t, 19.99, widget.Price)
	require.Equal(t, 100, widget.Stock)
	require.Greater(t, widget.ID, 0)

	widget2, err := table.One(ctx, tx, widget.ID)
	assert.NotSame(t, widget, widget2)
	assert.Equal(t, *widget, *widget2)

	widget2.Price = 24.99
	widget2.Stock = 80
	widget, _, err = table.Upd(ctx, tx, widget2)
	assert.NotSame(t, widget2, widget)
	assert.Equal(t, *widget2, *widget)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 1)

	found := false
	for _, r := range rows {
		if r.ID == widget.ID {
			found = true
			assert.Equal(t, 24.99, r.Price)
			assert.Equal(t, 80, r.Stock)
		}
	}
	require.True(t, found, "Widget should be in the list")

	_, err = table.Del(ctx, tx, widget.ID)
	require.NoError(t, err)

	_, err = table.One(ctx, tx, widget.ID)
	require.Error(t, err)
}
