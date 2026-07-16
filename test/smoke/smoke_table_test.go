//go:build smoke

package smoke

import (
	quick_crud "quick-crud"
	"quick-crud/dialect"
	"quick-crud/filter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTable_IdNameAgeRowFilled(t *testing.T) {
	t.Parallel()
	table := quick_crud.NewTable[IdNameAgeRowFilled](dialect.SQLiteDialect{})
	tableInfo := table.Internals().TableInfo
	assert.Equal(t, "id_name_age_filled", tableInfo.SQLName)
	require.Len(t, tableInfo.Fields, 3)
	assert.Equal(t, "id", tableInfo.Fields[0].SQLName)
	assert.Equal(t, "name", tableInfo.Fields[1].SQLName)
	assert.Equal(t, "age", tableInfo.Fields[2].SQLName)
}

func TestTable_IdNameAgeRowRow(t *testing.T) {
	t.Parallel()
	table := quick_crud.NewTable[IdNameAgeRow](dialect.SQLiteDialect{})
	tableInfo := table.Internals().TableInfo
	assert.Equal(t, "id_name_age_row", tableInfo.SQLName)
	require.Len(t, tableInfo.Fields, 3)
	assert.Equal(t, "id", tableInfo.Fields[0].SQLName)
	assert.Equal(t, "name", tableInfo.Fields[1].SQLName)
	assert.Equal(t, "age", tableInfo.Fields[2].SQLName)
}

var (
	unagedIdNameAge = &IdNameAgeRowFilled{ID: 1, Name: "Unaged", Age: nil}
	aliceIdNameAge  = &IdNameAgeRowFilled{ID: 2, Name: "Alice", Age: new(11)}
	bobIdNameAge    = &IdNameAgeRowFilled{ID: 3, Name: "Bob", Age: new(22)}
)

func TestTable_One(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := quick_crud.NewTable[IdNameAgeRowFilled](dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	row, err := table.One(env.ctx, tx, 1)
	require.NoError(t, err)
	assert.Equal(t, unagedIdNameAge, row)
}

func TestTable_Many(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := quick_crud.NewTable[IdNameAgeRowFilled](dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	rows, err := table.Many(env.ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 3)
	require.Equal(t, []*IdNameAgeRowFilled{aliceIdNameAge, bobIdNameAge, unagedIdNameAge}, rows)

	rows, err = table.Many(env.ctx, tx, &filter.Filter{
		Offset: 1,
		Limit:  2,
		Range:  nil,
	})
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	require.Equal(t, []*IdNameAgeRowFilled{bobIdNameAge, unagedIdNameAge}, rows)

	rows, err = table.Many(env.ctx, tx, &filter.Filter{
		Offset: 0,
		Limit:  0,
		Range: filter.ConditionNode{
			FieldIdx: 0,
			Op:       filter.CmdEq,
			Value:    2,
		},
	})
	require.NoError(t, err, table.Internals().SqlTexts.ListStart)
	assert.Len(t, rows, 1)
	require.Equal(t, []*IdNameAgeRowFilled{aliceIdNameAge}, rows)

}

func TestTable_AllCRUD(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := quick_crud.NewTable[IdNameAgeRow](dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	ctx := env.ctx
	colin, _, err := table.Ins(ctx, tx, &IdNameAgeRow{Name: "Colin", Age: 33})
	require.NoError(t, err)

	bob, _, err := table.Ins(ctx, tx, &IdNameAgeRow{Name: "Bob", Age: 22})
	require.NoError(t, err)

	alice, _, err := table.Ins(ctx, tx, &IdNameAgeRow{Name: "Alice", Age: 1})
	require.NoError(t, err)
	require.Equal(t, &IdNameAgeRow{ID: 3, Name: "Alice", Age: 1}, alice)

	alice2, err := table.One(ctx, tx, alice.ID)
	assert.NotSame(t, alice, alice2)
	assert.Equal(t, *alice, *alice2)

	alice2.Age = 11
	alice, _, err = table.Upd(ctx, tx, alice2)
	assert.NotSame(t, alice2, alice)
	assert.Equal(t, *alice2, *alice)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	require.Equal(t, []*IdNameAgeRow{alice, bob, colin}, rows)

	_, err = table.Del(ctx, tx, bob.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Equal(t, []*IdNameAgeRow{alice, colin}, rows)

	_, err = table.Del(ctx, tx, colin.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, []*IdNameAgeRow{alice}, rows)
}
