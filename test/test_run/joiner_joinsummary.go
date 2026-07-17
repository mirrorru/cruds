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

func JoinSummary_Reflection_OneMany(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner, err := cruds.NewJoiner[testmodels.JoinSummary](d)
	require.NoError(t, err)

	result, err := joiner.One(ctx, tx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	assert.Equal(t, testmodels.GenderTypeMale, result.From.Gender)
	require.NotNil(t, result.InnerVal)
	assert.Equal(t, "inner_1", result.InnerVal.InnerName)
	require.NotNil(t, result.LeftVal)
	assert.Equal(t, "left_1", result.LeftVal.LeftName)

	result2, err := joiner.One(ctx, tx, 2)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, "from_2", result2.From.Name)
	assert.Equal(t, testmodels.GenderTypeFemale, result2.From.Gender)
	assert.NotNil(t, result2.InnerVal)
	assert.Equal(t, "inner_2", result2.InnerVal.InnerName)
	assert.Nil(t, result2.LeftVal)

	_, err = joiner.One(ctx, tx, 999)
	require.Error(t, err)

	manyAll, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, manyAll, 2)
}

func JoinSummary_Typed_OneMany(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner := testmodels.NewJoinerJoinSummaryVal(d)

	result, err := joiner.One(ctx, tx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	assert.Equal(t, testmodels.GenderTypeMale, result.From.Gender)
	require.NotNil(t, result.InnerVal)
	assert.Equal(t, "inner_1", result.InnerVal.InnerName)
	require.NotNil(t, result.LeftVal)
	assert.Equal(t, "left_1", result.LeftVal.LeftName)

	result2, err := joiner.One(ctx, tx, 2)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, "from_2", result2.From.Name)
	assert.Nil(t, result2.LeftVal)

	_, err = joiner.One(ctx, tx, 999)
	require.Error(t, err)

	manyAll, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, manyAll, 2)
}
