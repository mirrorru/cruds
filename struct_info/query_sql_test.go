package struct_info_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mirrorru/crudquick/dialect"
	"github.com/mirrorru/crudquick/struct_info"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildQuerySqlTexts(t *testing.T) {
	t.Parallel()

	d := dialect.SQLiteDialect{}

	t.Run("single table (no JOINs)", func(t *testing.T) {
		type SingleTableQuery struct {
			User UserRow `tbl:"from"`
		}

		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(SingleTableQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		assert.Contains(t, sqlTexts.ListStart, "SELECT")
		assert.Contains(t, sqlTexts.ListStart, "FROM user_row")
		assert.NotContains(t, sqlTexts.ListStart, "JOIN")

		assert.Contains(t, sqlTexts.GetOne, "WHERE")
		assert.Contains(t, sqlTexts.GetOne, "user_row.id = ?")

		assert.Contains(t, sqlTexts.SortPart, "ORDER BY")
		assert.Contains(t, sqlTexts.SortPart, "user_row.name")
	})

	t.Run("two tables with LEFT JOIN", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// Check SELECT
		assert.Contains(t, sqlTexts.ListStart, "SELECT")
		assert.Contains(t, sqlTexts.ListStart, "user_row.id")
		assert.Contains(t, sqlTexts.ListStart, "user_row.name")
		assert.Contains(t, sqlTexts.ListStart, "c1.id")
		assert.Contains(t, sqlTexts.ListStart, "c1.user_id")
		assert.Contains(t, sqlTexts.ListStart, "c1.text")

		// Check FROM
		assert.Contains(t, sqlTexts.ListStart, "FROM user_row")

		// Check JOIN
		assert.Contains(t, sqlTexts.ListStart, "LEFT JOIN comment_row AS c1")
		assert.Contains(t, sqlTexts.ListStart, "ON user_row.id = c1.user_id")

		// Check WHERE (for GetOne)
		assert.Contains(t, sqlTexts.GetOne, "WHERE")
		assert.Contains(t, sqlTexts.GetOne, "user_row.id = ?")

		// Check ORDER BY
		assert.Contains(t, sqlTexts.SortPart, "ORDER BY user_row.name")
	})

	t.Run("multiple JOINs with aliases", func(t *testing.T) {
		type OrderRow struct {
			ID     int `tbl:"pk;auto"`
			UserID int `tbl:"ref=user_row:id"`
		}

		type MultiJoinQuery struct {
			User    UserRow    `tbl:"from;alias=u1"`
			Comment CommentRow `tbl:"join=left;alias=c1"`
			Order   OrderRow   `tbl:"join=inner;alias=o1"`
		}

		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(MultiJoinQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// Check FROM with alias
		assert.Contains(t, sqlTexts.ListStart, "FROM user_row AS u1")

		// Check JOINs
		assert.Contains(t, sqlTexts.ListStart, "LEFT JOIN comment_row AS c1")
		assert.Contains(t, sqlTexts.ListStart, "INNER JOIN order_row AS o1")

		// Check ON conditions
		assert.Contains(t, sqlTexts.ListStart, "ON u1.id = c1.user_id")
		assert.Contains(t, sqlTexts.ListStart, "ON u1.id = o1.user_id")
	})

	t.Run("WHERE clause with PK table", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(PKTableQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// PK table is Comment (index 1), so WHERE should use comment_row.id
		assert.Contains(t, sqlTexts.GetOne, "WHERE")
		assert.Contains(t, sqlTexts.GetOne, "comment_row.id = ?")
	})

	t.Run("ORDER BY with multiple fields", func(t *testing.T) {
		type SortedQuery struct {
			User    UserRow    `tbl:"from;sort=1"`
			Comment CommentRow `tbl:"join=left;alias=c1;sort=2"`
		}

		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(SortedQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// Should have ORDER BY with fields from both tables
		assert.Contains(t, sqlTexts.SortPart, "ORDER BY")
		// user_row.name has sort=1, c1 fields don't have sort, so only user_row.name
		assert.Contains(t, sqlTexts.SortPart, "user_row.name")
	})

	t.Run("no alias when same as table name", func(t *testing.T) {
		type NoAliasQuery struct {
			User UserRow `tbl:"from"`
		}

		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(NoAliasQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// Should not have "AS user_row" since alias equals table name
		assert.Contains(t, sqlTexts.ListStart, "FROM user_row")
		assert.NotContains(t, sqlTexts.ListStart, "AS user_row")
	})

	t.Run("alias differs from table name", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(TwoTablesQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// Should have "AS u1" since alias differs from table name
		assert.Contains(t, sqlTexts.ListStart, "FROM user_row AS u1")
		assert.Contains(t, sqlTexts.ListStart, "INNER JOIN comment_row AS c1")
	})

	t.Run("JOIN condition format", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// ON clause should be: <target_alias>.<target_col> = <source_alias>.<source_col>
		// For CommentRow.UserID with ref=user_row:id
		// target_alias = user_row, target_col = id
		// source_alias = c1, source_col = user_id
		assert.Contains(t, sqlTexts.ListStart, "ON user_row.id = c1.user_id")
	})

	t.Run("empty query info", func(t *testing.T) {
		qi := struct_info.QueryInfo{}
		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		assert.Empty(t, sqlTexts.GetOne)
		assert.Empty(t, sqlTexts.ListStart)
		assert.Empty(t, sqlTexts.SortPart)
	})

	t.Run("SQL structure validation", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQuery{}))
		require.NoError(t, err)

		sqlTexts := struct_info.BuildQuerySqlTexts(d, &qi)

		// Validate SQL structure
		assert.True(t, strings.HasPrefix(sqlTexts.ListStart, "SELECT"))
		assert.Contains(t, sqlTexts.ListStart, "FROM")
		assert.Contains(t, sqlTexts.ListStart, "LEFT JOIN")
		assert.Contains(t, sqlTexts.ListStart, "ON")

		assert.True(t, strings.HasPrefix(sqlTexts.GetOne, "SELECT"))
		assert.Contains(t, sqlTexts.GetOne, "WHERE")
		assert.Contains(t, sqlTexts.GetOne, "?")
	})
}
