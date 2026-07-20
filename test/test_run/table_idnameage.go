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

const IdNameAgeRowFilledCRUDTable = "id_name_age_filled_crud"

type idNameAgeRowFilledCRUD struct {
	testmodels.IdNameAgeRowFilled
}

func (idNameAgeRowFilledCRUD) SQLName() string { return IdNameAgeRowFilledCRUDTable }

func IdNameAgeRowFilled_Reflection_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[idNameAgeRowFilledCRUD](d)

	age11 := 11
	age22 := 22

	unaged, _, err := table.Ins(ctx, tx, &idNameAgeRowFilledCRUD{IdNameAgeRowFilled: testmodels.IdNameAgeRowFilled{ID: 1, Name: "Unaged", Age: nil}})
	require.NoError(t, err)
	assert.Equal(t, "Unaged", unaged.Name)
	assert.Nil(t, unaged.Age)

	alice, _, err := table.Ins(ctx, tx, &idNameAgeRowFilledCRUD{IdNameAgeRowFilled: testmodels.IdNameAgeRowFilled{ID: 2, Name: "Alice", Age: &age11}})
	require.NoError(t, err)
	assert.Equal(t, "Alice", alice.Name)

	bob, _, err := table.Ins(ctx, tx, &idNameAgeRowFilledCRUD{IdNameAgeRowFilled: testmodels.IdNameAgeRowFilled{ID: 3, Name: "Bob", Age: &age22}})
	require.NoError(t, err)

	found, err := table.One(ctx, tx, unaged.ID)
	require.NoError(t, err)
	assert.Equal(t, unaged, found)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 3)

	_, err = table.Del(ctx, tx, bob.ID)
	require.NoError(t, err)

	_, err = table.Del(ctx, tx, alice.ID)
	require.NoError(t, err)

	_, err = table.Del(ctx, tx, unaged.ID)
	require.NoError(t, err)
}
