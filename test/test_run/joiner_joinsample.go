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

func JoinSample_Reflection_OneMany(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner, err := cruds.NewJoiner[testmodels.JoinSample](d)
	require.NoError(t, err)

	result, err := joiner.One(ctx, tx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	assert.Equal(t, int64(10), result.Inner.ID)
	assert.Equal(t, "inner_1", result.Inner.InnerVal)
	require.NotNil(t, result.Left)
	assert.Equal(t, "left_1", result.Left.LeftVal)

	result2, err := joiner.One(ctx, tx, 2)
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.Equal(t, "from_2", result2.From.Name)
	assert.Nil(t, result2.Left)

	_, err = joiner.One(ctx, tx, 999)
	require.Error(t, err)

	manyAll, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, manyAll, 2)
}

func JoinSample_Reflection_NullPointer(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner, err := cruds.NewJoiner[testmodels.JoinSample](d)
	require.NoError(t, err)

	results, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	var res1, res2, res3 *testmodels.JoinSample
	for _, r := range results {
		switch r.From.ID {
		case 1:
			res1 = r
		case 2:
			res2 = r
		case 3:
			res3 = r
		}
	}
	require.NotNil(t, res1.Left)
	assert.Equal(t, "left_1", res1.Left.LeftVal)
	assert.Nil(t, res2.Left)
	require.NotNil(t, res3.Left)
	assert.Equal(t, "left_3", res3.Left.LeftVal)

	assert.NotSame(t, res1.Left, res3.Left)

	originalVal := res3.Left.LeftVal
	res1.Left.LeftVal = "modified"
	assert.Equal(t, originalVal, res3.Left.LeftVal)
}

func JoinSample_Typed_OneMany(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner := testmodels.NewJoinerJoinSampleVal(d)

	result, err := joiner.One(ctx, tx, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	assert.Equal(t, "inner_1", result.Inner.InnerVal)
	require.NotNil(t, result.Left)
	assert.Equal(t, "left_1", result.Left.LeftVal)

	result2, err := joiner.One(ctx, tx, 2)
	require.NoError(t, err)
	assert.Equal(t, "from_2", result2.From.Name)
	assert.Nil(t, result2.Left)

	_, err = joiner.One(ctx, tx, 999)
	require.Error(t, err)

	manyAll, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, manyAll, 2)
}

func JoinSample_Typed_NullPointer(t *testing.T, tx cruds.TxProcessor, d dialect.SQLDialect) {
	t.Helper()
	ctx := context.Background()

	joiner := testmodels.NewJoinerJoinSampleVal(d)

	results, err := joiner.Many(ctx, tx, nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	var res1, res2, res3 *testmodels.JoinSample
	for _, r := range results {
		switch r.From.ID {
		case 1:
			res1 = r
		case 2:
			res2 = r
		case 3:
			res3 = r
		}
	}
	require.NotNil(t, res1.Left)
	assert.Equal(t, "left_1", res1.Left.LeftVal)
	assert.Nil(t, res2.Left)
	require.NotNil(t, res3.Left)
	assert.Equal(t, "left_3", res3.Left.LeftVal)

	assert.NotSame(t, res1.Left, res3.Left)

	originalVal := res3.Left.LeftVal
	res1.Left.LeftVal = "modified"
	assert.Equal(t, originalVal, res3.Left.LeftVal)
}
