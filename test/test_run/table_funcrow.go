//go:build smoke || functional

package test_run

import (
	"context"
	"testing"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/dialect"
	"github.com/mirrorru/cruds/test/testmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const FuncRowCRUDTable = "func_row_crud"
const FuncRowTypedCRUDTable = "func_row_typed_crud"

type funcRowCRUD struct {
	testmodels.FuncRow
}

func (funcRowCRUD) SQLName() string { return FuncRowCRUDTable }

type funcRowTypedCRUD struct {
	testmodels.FuncRow
}

func (funcRowTypedCRUD) SQLName() string { return FuncRowTypedCRUDTable }

func FuncRow_Reflection_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[funcRowCRUD](d)

	row1, _, err := table.Ins(ctx, tx, &funcRowCRUD{FuncRow: testmodels.FuncRow{Name: "Alpha", Value: 100}})
	require.NoError(t, err)
	assert.Equal(t, "Alpha", row1.Name)
	assert.Equal(t, 100, row1.Value)
	assert.NotZero(t, row1.ID)

	found, err := table.One(ctx, tx, row1.ID)
	require.NoError(t, err)
	assert.Equal(t, row1.ID, found.ID)
	assert.Equal(t, "Alpha", found.Name)

	upd, _, err := table.Upd(ctx, tx, &funcRowCRUD{FuncRow: testmodels.FuncRow{ID: row1.ID, Name: "AlphaUpdated", Value: 150}})
	require.NoError(t, err)
	assert.Equal(t, "AlphaUpdated", upd.Name)
	assert.Equal(t, 150, upd.Value)

	_, _, err = table.Ins(ctx, tx, &funcRowCRUD{FuncRow: testmodels.FuncRow{Name: "Beta", Value: 200}})
	require.NoError(t, err)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	result, err := table.Del(ctx, tx, row1.ID)
	require.NoError(t, err)
	affected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 1)

	_, err = table.One(ctx, tx, row1.ID)
	assert.Error(t, err)
}

func FuncRow_Typed_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[funcRowTypedCRUD](d)

	row1, _, err := table.Ins(ctx, tx, &funcRowTypedCRUD{FuncRow: testmodels.FuncRow{Name: "TypedAlpha", Value: 10}})
	require.NoError(t, err)
	assert.Equal(t, "TypedAlpha", row1.Name)
	assert.Equal(t, 10, row1.Value)
	assert.NotZero(t, row1.ID)

	found, err := table.One(ctx, tx, row1.ID)
	require.NoError(t, err)
	assert.Equal(t, row1.ID, found.ID)
	assert.Equal(t, "TypedAlpha", found.Name)

	_, _, err = table.Ins(ctx, tx, &funcRowTypedCRUD{FuncRow: testmodels.FuncRow{Name: "TypedBeta", Value: 20}})
	require.NoError(t, err)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	upd, _, err := table.Upd(ctx, tx, &funcRowTypedCRUD{FuncRow: testmodels.FuncRow{ID: row1.ID, Name: "TypedAlphaUpdated", Value: 15}})
	require.NoError(t, err)
	assert.Equal(t, "TypedAlphaUpdated", upd.Name)
	assert.Equal(t, 15, upd.Value)

	result, err := table.Del(ctx, tx, row1.ID)
	require.NoError(t, err)
	affected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 1)

	_, err = table.One(ctx, tx, row1.ID)
	assert.Error(t, err)
}
