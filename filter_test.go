package quick_crud_test

import (
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
	f := qc.Filter{Range: nil}
	query, args, err := f.BuildWhere(table.Internals().TableInfo.Fields, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Empty(t, query)
	assert.Nil(t, args)
}

func TestBuildWhere_SingleCondition_SQLite(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := qc.Cond(1, qc.CmdEq, "Alice")
	f := qc.Filter{Range: root}

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

	root := qc.Cond(1, qc.CmdEq, "Alice")
	f := qc.Filter{Range: root}

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

	root := qc.And(
		qc.Cond(1, qc.CmdEq, "Alice"),
		qc.Cond(2, qc.CmdGte, 25),
	)
	f := qc.Filter{Range: root}

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

	root := qc.Or(
		qc.Cond(1, qc.CmdEq, "Alice"),
		qc.Cond(1, qc.CmdEq, "Bob"),
	)
	f := qc.Filter{Range: root}

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

	root := qc.Not(qc.Cond(2, qc.CmdEq, 18))
	f := qc.Filter{Range: root}

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

	root := qc.And(
		qc.Cond(1, qc.CmdEq, "Alice"),
		qc.Or(
			qc.Cond(2, qc.CmdGt, 20),
			qc.Cond(2, qc.CmdLt, 10),
		),
	)
	f := qc.Filter{Range: root}

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
		op       qc.CommandOp
		value    any
		wantSQL  string
		wantArgs []any
	}{
		{"Eq", qc.CmdEq, 42, "= ?", []any{42}},
		{"NotEq", qc.CmdNotEq, 42, "<> ?", []any{42}},
		{"Gt", qc.CmdGt, 42, "> ?", []any{42}},
		{"Gte", qc.CmdGte, 42, ">= ?", []any{42}},
		{"Lt", qc.CmdLt, 42, "< ?", []any{42}},
		{"Lte", qc.CmdLte, 42, "<= ?", []any{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
			tf := table.Internals().TableInfo.Fields

			root := qc.Cond(2, tt.op, tt.value)
			f := qc.Filter{Range: root}

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

		root := qc.Cond(1, qc.CmdIsNull, nil)
		f := qc.Filter{Range: root}

		query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
		require.NoError(t, err)
		assert.Contains(t, query, "IS NULL")
		assert.Nil(t, args)
	})

	t.Run("IsNotNull", func(t *testing.T) {
		t.Parallel()

		table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
		tf := table.Internals().TableInfo.Fields

		root := qc.Cond(1, qc.CmdIsNotNull, nil)
		f := qc.Filter{Range: root}

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

	root := qc.Cond(1, qc.CmdLike, "%Alice%")
	f := qc.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "LIKE ?")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_ILike_PG(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.PostgreSQLDialect{})
	tf := table.Internals().TableInfo.Fields

	root := qc.Cond(1, qc.CmdILike, "%Alice%")
	f := qc.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.PostgreSQLDialect{})
	require.NoError(t, err)
	assert.Contains(t, query, "ILIKE")
	assert.Equal(t, []any{"%Alice%"}, args)
}

func TestBuildWhere_ILike_SQLite(t *testing.T) {
	t.Parallel()

	table := qc.NewTable[filterTestRow](dialect.SQLiteDialect{})
	tf := table.Internals().TableInfo.Fields

	root := qc.Cond(1, qc.CmdILike, "%Alice%")
	f := qc.Filter{Range: root}

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

	root := qc.Cond(2, qc.CmdIn, []any{20, 30, 40})
	f := qc.Filter{Range: root}

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

	root := qc.And(
		qc.Cond(1, qc.CmdEq, "Alice"),
		qc.Cond(2, qc.CmdGte, 25),
		qc.Cond(2, qc.CmdLte, 50),
	)
	f := qc.Filter{Range: root}

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

	root := qc.Cond(99, qc.CmdEq, "test")
	f := qc.Filter{Range: root}

	query, args, err := f.BuildWhere(tf, dialect.SQLiteDialect{})
	require.Error(t, err)
	assert.Empty(t, query)
	assert.Nil(t, args)
}

func TestCond_HasRequiredMethods(t *testing.T) {
	t.Parallel()

	node := qc.Cond(0, qc.CmdEq, "test")
	assert.NotNil(t, node)

	andNode := qc.And(node)
	assert.NotNil(t, andNode)

	orNode := qc.Or(node, node)
	assert.NotNil(t, orNode)

	notNode := qc.Not(node)
	require.NotNil(t, notNode)
}
