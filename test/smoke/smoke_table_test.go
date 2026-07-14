//go:build smoke

package smoke

import (
	quick_crud "quick-crud"
	"quick-crud/dialect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTable_Internals(t *testing.T) {
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
	unagedIdNameAge = &IdNameAgeRow{ID: 1, Name: "Unaged", Age: nil}
	aliceIdNameAge  = &IdNameAgeRow{ID: 2, Name: "Alice", Age: new(11)}
	bobIdNameAge    = &IdNameAgeRow{ID: 3, Name: "Bob", Age: new(22)}
)

func TestTable_One(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := quick_crud.NewTable[IdNameAgeRow](dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	row, err := table.One(env.ctx, tx, 1)
	require.NoError(t, err)
	assert.Equal(t, unagedIdNameAge, row)
}

func TestTable_Many(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	table := quick_crud.NewTable[IdNameAgeRow](dialect.SQLiteDialect{})
	tx := env.TxProcessor()
	rows, err := table.Many(env.ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 3)
	require.Equal(t, []*IdNameAgeRow{unagedIdNameAge, aliceIdNameAge, bobIdNameAge}, rows)

	rows, err = table.Many(env.ctx, tx, &quick_crud.Filter{
		Offset: 1,
		Limit:  2,
		Range:  nil,
	})
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	require.Equal(t, []*IdNameAgeRow{aliceIdNameAge, bobIdNameAge}, rows)

	rows, err = table.Many(env.ctx, tx, &quick_crud.Filter{
		Offset: 0,
		Limit:  0,
		Range: quick_crud.ConditionNode{
			FieldIdx: 0,
			Op:       quick_crud.CmdEq,
			Value:    2,
		},
	})
	require.NoError(t, err, table.Internals().SqlTexts.ListStart)
	assert.Len(t, rows, 1)
	require.Equal(t, []*IdNameAgeRow{aliceIdNameAge}, rows)

}
