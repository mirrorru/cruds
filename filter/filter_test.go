package filter_test

import (
	"quick-crud/filter"
	"testing"

	qc "quick-crud"
	"quick-crud/dialect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type filterTestRow struct {
	ID   int64  `tbl:"pk;auto"`
	Name string `tbl:"col=user_name"`
	Age  int
}

func TestBuildWhere_NilRoot(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	f := filter.Filter{Range: nil}
	query, args, err := f.BuildWhere(table.Internals().TableInfo.Fields, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Empty(t, query)
	assert.Nil(t, args)
}

func TestBuildWhere_SingleCondition_SQLite(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(1, filter.CmdEq, "Alice")
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "user_name")
	assert.Contains(t, query, "= ?")
	assert.Equal(t, []any{"Alice"}, args)
}

func TestBuildWhere_SingleCondition_PG(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(1, filter.CmdEq, "Alice")
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.PostgreSQLDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "user_name")
	assert.Contains(t, query, "= $1")
	assert.Equal(t, []any{"Alice"}, args)
}

func TestBuildWhere_MultipleAND(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.And(
		filter.Cond(1, filter.CmdEq, "Alice"),
		filter.Cond(2, filter.CmdGte, 25),
	)
	f := filter.Filter{Range: root}

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

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Or(
		filter.Cond(1, filter.CmdEq, "Alice"),
		filter.Cond(1, filter.CmdEq, "Bob"),
	)
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, " OR ")
	assert.Equal(t, []any{"Alice", "Bob"}, args)
}

func TestBuildWhere_NOT(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Not(filter.Cond(2, filter.CmdEq, 18))
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "NOT")
	assert.Contains(t, query, "age")
	assert.Equal(t, []any{18}, args)
}

func TestBuildWhere_Nested(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.And(
		filter.Cond(1, filter.CmdEq, "Alice"),
		filter.Or(
			filter.Cond(2, filter.CmdGt, 20),
			filter.Cond(2, filter.CmdLt, 10),
		),
	)
	f := filter.Filter{Range: root}

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
		op       filter.CommandOp
		value    any
		wantSQL  string
		wantArgs []any
	}{
		{"Eq", filter.CmdEq, 42, "= ?", []any{42}},
		{"NotEq", filter.CmdNotEq, 42, "<> ?", []any{42}},
		{"Gt", filter.CmdGt, 42, "> ?", []any{42}},
		{"Gte", filter.CmdGte, 42, ">= ?", []any{42}},
		{"Lt", filter.CmdLt, 42, "< ?", []any{42}},
		{"Lte", filter.CmdLte, 42, "<= ?", []any{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
			tf := table.Internals().TableInfo.Fields

			root := filter.Cond(2, tt.op, tt.value)
			f := filter.Filter{Range: root}

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

		table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
		tf := table.Internals().TableInfo.Fields

		root := filter.Cond(1, filter.CmdIsNull, nil)
		f := filter.Filter{Range: root}

		query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
		require.NoError(t, err)
		assert.Contains(t, query, "IS NULL")
		assert.Nil(t, args)
	})

	t.Run("IsNotNull", func(t *testing.T) {
		t.Parallel()

		table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
		tf := table.Internals().TableInfo.Fields

		root := filter.Cond(1, filter.CmdIsNotNull, nil)
		f := filter.Filter{Range: root}

		query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
		require.NoError(t, err)
		assert.Contains(t, query, "IS NOT NULL")
		assert.Nil(t, args)
	})
}

func TestBuildWhere_Like(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(1, filter.CmdLike, "%Alice%")
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "LIKE ?")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_ILike_PG(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(1, filter.CmdILike, "%Alice%")
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.PostgreSQLDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "ILIKE")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_ILike_SQLite(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(1, filter.CmdILike, "%Alice%")
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "LOWER(")
	assert.Contains(t, query, ") LIKE LOWER(")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_In(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(2, filter.CmdIn, []any{20, 30, 40})
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "IN")
	assert.Contains(t, query, "?, ?, ?")
	assert.Equal(t, []any{20, 30, 40}, args)
}

func TestBuildWhere_PlaceholderContinuity(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.And(
		filter.Cond(1, filter.CmdEq, "Alice"),
		filter.Cond(2, filter.CmdGte, 25),
		filter.Cond(2, filter.CmdLte, 50),
	)
	f := filter.Filter{Range: root}

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

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := filter.Cond(99, filter.CmdEq, "test")
	f := filter.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.Error(t, err)
	assert.Empty(t, query)
	assert.Nil(t, args)
}

func TestCond_HasRequiredMethods(t *testing.T) {
	t.Parallel()

	node := filter.Cond(0, filter.CmdEq, "test")
	assert.NotNil(t, node)

	andNode := filter.And(node)
	assert.NotNil(t, andNode)

	orNode := filter.Or(node, node)
	assert.NotNil(t, orNode)

	notNode := filter.Not(node)
	require.NotNil(t, notNode)
}
