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

const UserRowCRUDTable = "user_row_crud"
const UserRowTypedCRUDTable = "user_row_typed_crud"

type userRowCRUD struct {
	testmodels.UserRow
}

func (userRowCRUD) SQLName() string { return UserRowCRUDTable }

type userRowTypedCRUD struct {
	testmodels.UserRow
}

func (userRowTypedCRUD) SQLName() string { return UserRowTypedCRUDTable }

func UserRow_Reflection_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[userRowCRUD](d)

	alice, _, err := table.Ins(ctx, tx, &userRowCRUD{UserRow: testmodels.UserRow{Name: "Alice", Age: 30}})
	require.NoError(t, err)
	assert.Equal(t, "Alice", alice.Name)
	assert.Equal(t, 30, alice.Age)

	found, err := table.One(ctx, tx, alice.ID)
	require.NoError(t, err)
	assert.Equal(t, alice.ID, found.ID)
	assert.Equal(t, "Alice", found.Name)

	upd, _, err := table.Upd(ctx, tx, &userRowCRUD{UserRow: testmodels.UserRow{ID: alice.ID, Name: "AliceUpdated", Age: 31}})
	require.NoError(t, err)
	assert.Equal(t, "AliceUpdated", upd.Name)
	assert.Equal(t, 31, upd.Age)

	bob, _, err := table.Ins(ctx, tx, &userRowCRUD{UserRow: testmodels.UserRow{Name: "Bob", Age: 22}})
	require.NoError(t, err)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	_, err = table.Del(ctx, tx, bob.ID)
	require.NoError(t, err)

	_, err = table.Del(ctx, tx, alice.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 0)
}

func UserRow_Typed_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[userRowTypedCRUD](d)

	alice, _, err := table.Ins(ctx, tx, &userRowTypedCRUD{UserRow: testmodels.UserRow{Name: "Alice", Age: 30}})
	require.NoError(t, err)
	assert.Equal(t, "Alice", alice.Name)
	assert.Equal(t, 30, alice.Age)

	found, err := table.One(ctx, tx, alice.ID)
	require.NoError(t, err)
	assert.Equal(t, alice.ID, found.ID)
	assert.Equal(t, "Alice", found.Name)

	upd, _, err := table.Upd(ctx, tx, &userRowTypedCRUD{UserRow: testmodels.UserRow{ID: alice.ID, Name: "AliceUpdated", Age: 31}})
	require.NoError(t, err)
	assert.Equal(t, "AliceUpdated", upd.Name)
	assert.Equal(t, 31, upd.Age)

	bob, _, err := table.Ins(ctx, tx, &userRowTypedCRUD{UserRow: testmodels.UserRow{Name: "Bob", Age: 22}})
	require.NoError(t, err)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	_, err = table.Del(ctx, tx, bob.ID)
	require.NoError(t, err)

	_, err = table.Del(ctx, tx, alice.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 0)
}
