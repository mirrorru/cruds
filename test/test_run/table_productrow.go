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

const ProductRowCRUDTable = "products_crud"
const ProductRowTypedCRUDTable = "products_typed_crud"

type productRowCRUD struct {
	testmodels.ProductRow
}

func (productRowCRUD) SQLName() string { return ProductRowCRUDTable }

type productRowTypedCRUD struct {
	testmodels.ProductRow
}

func (productRowTypedCRUD) SQLName() string { return ProductRowTypedCRUDTable }

func ProductRow_Reflection_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[productRowCRUD](d)

	widget, _, err := table.Ins(ctx, tx, &productRowCRUD{ProductRow: testmodels.ProductRow{Name: "Widget", Price: 19.99, Stock: 100}})
	require.NoError(t, err)
	assert.Equal(t, "Widget", widget.Name)
	assert.InDelta(t, 19.99, widget.Price, 0.001)
	assert.Equal(t, 100, widget.Stock)

	found, err := table.One(ctx, tx, widget.ID)
	require.NoError(t, err)
	assert.Equal(t, widget.ID, found.ID)
	assert.Equal(t, "Widget", found.Name)

	upd, _, err := table.Upd(ctx, tx, &productRowCRUD{ProductRow: testmodels.ProductRow{ID: widget.ID, Name: "UpdatedWidget", Price: 29.99, Stock: 80}})
	require.NoError(t, err)
	assert.Equal(t, "UpdatedWidget", upd.Name)
	assert.InDelta(t, 29.99, upd.Price, 0.001)
	assert.Equal(t, 80, upd.Stock)

	gadget, _, err := table.Ins(ctx, tx, &productRowCRUD{ProductRow: testmodels.ProductRow{Name: "Gadget", Price: 9.99, Stock: 50}})
	require.NoError(t, err)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	_, err = table.Del(ctx, tx, gadget.ID)
	require.NoError(t, err)
	_, err = table.Del(ctx, tx, widget.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 0)
}

func ProductRow_Typed_CRUD(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()
	table := cruds.NewTable[productRowTypedCRUD](d)

	widget, _, err := table.Ins(ctx, tx, &productRowTypedCRUD{ProductRow: testmodels.ProductRow{Name: "Widget", Price: 19.99, Stock: 100}})
	require.NoError(t, err)
	assert.Equal(t, "Widget", widget.Name)
	assert.InDelta(t, 19.99, widget.Price, 0.001)
	assert.Equal(t, 100, widget.Stock)

	found, err := table.One(ctx, tx, widget.ID)
	require.NoError(t, err)
	assert.Equal(t, widget.ID, found.ID)

	upd, _, err := table.Upd(ctx, tx, &productRowTypedCRUD{ProductRow: testmodels.ProductRow{ID: widget.ID, Name: "UpdatedWidget", Price: 29.99, Stock: 80}})
	require.NoError(t, err)
	assert.Equal(t, "UpdatedWidget", upd.Name)
	assert.InDelta(t, 29.99, upd.Price, 0.001)

	gadget, _, err := table.Ins(ctx, tx, &productRowTypedCRUD{ProductRow: testmodels.ProductRow{Name: "Gadget", Price: 9.99, Stock: 50}})
	require.NoError(t, err)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 2)

	_, err = table.Del(ctx, tx, gadget.ID)
	require.NoError(t, err)
	_, err = table.Del(ctx, tx, widget.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, rows, 0)
}
