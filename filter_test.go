package cruds_test

import (
	"testing"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/dialect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type filterTestRow struct {
	ID   int64  `crud:"pk;auto"`
	Name string `crud:"col=user_name"`
	Age  int
}

func TestBuildWhere_NilRoot(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	f := cruds.Filter{Range: nil}
	query, args, err := f.BuildWhere(table.Internals().TableInfo.Fields, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Empty(t, query)
	assert.Nil(t, args)
}

func TestBuildWhere_SingleCondition_SQLite(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(1, cruds.CmdEq, "Alice")
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "user_name")
	assert.Contains(t, query, "= ?")
	assert.Equal(t, []any{"Alice"}, args)
}

func TestBuildWhere_SingleCondition_PG(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(1, cruds.CmdEq, "Alice")
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.PostgreSQLDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "user_name")
	assert.Contains(t, query, "= $1")
	assert.Equal(t, []any{"Alice"}, args)
}

func TestBuildWhere_MultipleAND(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.And(
		cruds.Cond(1, cruds.CmdEq, "Alice"),
		cruds.Cond(2, cruds.CmdGte, 25),
	)
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "user_name")
	assert.Contains(t, query, "age")
	assert.Contains(t, query, " AND ")
	assert.Equal(t, []any{"Alice", 25}, args)
}

func TestBuildWhere_OR(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Or(
		cruds.Cond(1, cruds.CmdEq, "Alice"),
		cruds.Cond(1, cruds.CmdEq, "Bob"),
	)
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, " OR ")
	assert.Equal(t, []any{"Alice", "Bob"}, args)
}

func TestBuildWhere_NOT(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Not(cruds.Cond(2, cruds.CmdEq, 18))
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "NOT")
	assert.Contains(t, query, "age")
	assert.Equal(t, []any{18}, args)
}

func TestBuildWhere_Nested(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.And(
		cruds.Cond(1, cruds.CmdEq, "Alice"),
		cruds.Or(
			cruds.Cond(2, cruds.CmdGt, 20),
			cruds.Cond(2, cruds.CmdLt, 10),
		),
	)
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "user_name")
	assert.Contains(t, query, "age")
	assert.Contains(t, query, " AND ")
	assert.Contains(t, query, " OR ")
	assert.Equal(t, []any{"Alice", 20, 10}, args)
}

func TestBuildWhere_Operators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		op       cruds.CommandOp
		value    any
		wantSQL  string
		wantArgs []any
	}{
		{"Eq", cruds.CmdEq, 42, "= ?", []any{42}},
		{"NotEq", cruds.CmdNotEq, 42, "<> ?", []any{42}},
		{"Gt", cruds.CmdGt, 42, "> ?", []any{42}},
		{"Gte", cruds.CmdGte, 42, ">= ?", []any{42}},
		{"Lt", cruds.CmdLt, 42, "< ?", []any{42}},
		{"Lte", cruds.CmdLte, 42, "<= ?", []any{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
			tf := table.Internals().TableInfo.Fields

			root := cruds.Cond(2, tt.op, tt.value)
			f := cruds.Filter{Range: root}

			query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
			require.NoError(t, err)
			assert.Contains(t, query, tt.wantSQL)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestBuildWhere_IsNull_IsNotNull(t *testing.T) {
	t.Parallel()

	t.Run("IsNull", func(t *testing.T) {
		t.Parallel()

		table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
		tf := table.Internals().TableInfo.Fields

		root := cruds.Cond(1, cruds.CmdIsNull, nil)
		f := cruds.Filter{Range: root}

		query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
		require.NoError(t, err)
		assert.Contains(t, query, "IS NULL")
		assert.Nil(t, args)
	})

	t.Run("IsNotNull", func(t *testing.T) {
		t.Parallel()

		table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
		tf := table.Internals().TableInfo.Fields

		root := cruds.Cond(1, cruds.CmdIsNotNull, nil)
		f := cruds.Filter{Range: root}

		query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
		require.NoError(t, err)
		assert.Contains(t, query, "IS NOT NULL")
		assert.Nil(t, args)
	})
}

func TestBuildWhere_Like(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(1, cruds.CmdLike, "%Alice%")
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "LIKE ?")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_ILike_PG(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(1, cruds.CmdILike, "%Alice%")
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.PostgreSQLDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "ILIKE")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_ILike_SQLite(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(1, cruds.CmdILike, "%Alice%")
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "LOWER(")
	assert.Contains(t, query, ") LIKE LOWER(")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_In(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(2, cruds.CmdIn, []any{20, 30, 40})
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "IN")
	assert.Contains(t, query, "?, ?, ?")
	assert.Equal(t, []any{20, 30, 40}, args)
}

func TestBuildWhere_PlaceholderContinuity(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.And(
		cruds.Cond(1, cruds.CmdEq, "Alice"),
		cruds.Cond(2, cruds.CmdGte, 25),
		cruds.Cond(2, cruds.CmdLte, 50),
	)
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.PostgreSQLDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "$1")
	assert.Contains(t, query, "$2")
	assert.Contains(t, query, "$3")
	assert.NotContains(t, query, "$4")
	assert.Len(t, args, 3)
	assert.Equal(t, []any{"Alice", 25, 50}, args)
}

func TestBuildWhere_OutOfRange(t *testing.T) {
	t.Parallel()

	table := cruds.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := cruds.Cond(99, cruds.CmdEq, "test")
	f := cruds.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.Error(t, err)
	assert.Empty(t, query)
	assert.Nil(t, args)
}

func TestCond_HasRequiredMethods(t *testing.T) {
	t.Parallel()

	node := cruds.Cond(0, cruds.CmdEq, "test")
	assert.NotNil(t, node)

	andNode := cruds.And(node)
	assert.NotNil(t, andNode)

	orNode := cruds.Or(node, node)
	assert.NotNil(t, orNode)

	notNode := cruds.Not(node)
	require.NotNil(t, notNode)
}
