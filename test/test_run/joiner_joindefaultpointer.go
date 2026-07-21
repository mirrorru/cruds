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

func JoinDefaultPointer_Reflection_LeftJoinDefault(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner, err := cruds.NewJoiner[testmodels.JoinDefaultPointer](d)
	require.NoError(t, err)

	result, err := joiner.One(ctx, tx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	require.NotNil(t, result.Left)
	assert.Equal(t, "ptr_left_1", result.Left.Value)

	result2, err := joiner.One(ctx, tx, 2)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, "from_2", result2.From.Name)
	assert.Nil(t, result2.Left)

	manyAll, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, manyAll, 2)
}

func JoinDefaultPointer_Typed_LeftJoinDefault(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner := testmodels.NewJoinerJoinDefaultPointerVal(d)

	result, err := joiner.One(ctx, tx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	require.NotNil(t, result.Left)
	assert.Equal(t, "ptr_left_1", result.Left.Value)

	result2, err := joiner.One(ctx, tx, 2)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, "from_2", result2.From.Name)
	assert.Nil(t, result2.Left)

	manyAll, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, manyAll, 2)
}
