package struct_info_test

import (
	"reflect"
	"testing"

	"github.com/mirrorru/crudquick/struct_info"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserRow struct {
	ID   int    `tbl:"pk;auto"`
	Name string `tbl:"sort=1"`
}

type CommentRow struct {
	ID     int `tbl:"pk;auto"`
	UserID int `tbl:"ref=user_row:id"`
	Text   string
}

type UserCommentsQuery struct {
	User    UserRow    `tbl:"from"`
	Comment CommentRow `tbl:"join=left;alias=c1"`
}

type UserCommentsQueryPtr struct {
	User    UserRow     `tbl:"from"`
	Comment *CommentRow `tbl:"join=left;alias=c1"`
}

type TwoTablesQuery struct {
	First  UserRow    `tbl:"from;alias=u1"`
	Second CommentRow `tbl:"join=inner;alias=c1"`
}

type OmitFieldQuery struct {
	User    UserRow    `tbl:"from"`
	Comment CommentRow `tbl:"omit"`
}

type PKTableQuery struct {
	User    UserRow    `tbl:"from"`
	Comment CommentRow `tbl:"join=left;pk"`
}

func TestCollectQueryInfo(t *testing.T) {
	t.Parallel()

	t.Run("basic two tables", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQuery{}))
		require.NoError(t, err)

		assert.Equal(t, 2, len(qi.Tables))
		assert.Equal(t, 0, qi.FromIdx)
		assert.Equal(t, 0, qi.PKIdx)

		// Check FROM table
		assert.True(t, qi.Tables[0].IsFrom)
		assert.Equal(t, "user_row", qi.Tables[0].Alias)
		assert.False(t, qi.Tables[0].IsPointer)

		// Check JOIN table
		assert.False(t, qi.Tables[1].IsFrom)
		assert.Equal(t, "c1", qi.Tables[1].Alias)
		assert.Equal(t, "left", qi.Tables[1].JoinType)
		assert.False(t, qi.Tables[1].IsPointer)

		// Check combined fields
		assert.Equal(t, 5, len(qi.CombinedFields)) // user_row: id, name; c1: id, user_id, text
		assert.Equal(t, "user_row.id", qi.CombinedFields[0].SQLName)
		assert.Equal(t, "user_row.name", qi.CombinedFields[1].SQLName)
		assert.Equal(t, "c1.id", qi.CombinedFields[2].SQLName)
		assert.Equal(t, "c1.user_id", qi.CombinedFields[3].SQLName)
		assert.Equal(t, "c1.text", qi.CombinedFields[4].SQLName)

		// Check JOIN conditions
		assert.Equal(t, 1, len(qi.Tables[1].JoinConds))
		assert.Equal(t, "user_row", qi.Tables[1].JoinConds[0].TargetAlias)
		assert.Equal(t, "id", qi.Tables[1].JoinConds[0].TargetColumn)
		assert.Equal(t, "user_id", qi.Tables[1].JoinConds[0].SourceColumn)
	})

	t.Run("pointer T-field", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQueryPtr{}))
		require.NoError(t, err)

		assert.True(t, qi.Tables[1].IsPointer)
	})

	t.Run("auto FROM (first non-omit)", func(t *testing.T) {
		type AutoFromQuery struct {
			First  UserRow
			Second CommentRow `tbl:"from"`
		}

		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(AutoFromQuery{}))
		require.NoError(t, err)

		assert.Equal(t, 1, qi.FromIdx)
		assert.True(t, qi.Tables[1].IsFrom)
	})

	t.Run("explicit PK table", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(PKTableQuery{}))
		require.NoError(t, err)

		assert.Equal(t, 1, qi.PKIdx)
		assert.True(t, qi.Tables[1].IsPK)
	})

	t.Run("omit field", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(OmitFieldQuery{}))
		require.NoError(t, err)

		assert.Equal(t, 1, len(qi.Tables))
		assert.Equal(t, "user_row", qi.Tables[0].Alias)
	})

	t.Run("default JOIN type (inner for non-pointer)", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(TwoTablesQuery{}))
		require.NoError(t, err)

		assert.Equal(t, "inner", qi.Tables[1].JoinType)
	})

	t.Run("default JOIN type (left when FROM is pointer)", func(t *testing.T) {
		type PtrFromQuery struct {
			User    *UserRow   `tbl:"from"`
			Comment CommentRow `tbl:"alias=c1"`
		}

		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(PtrFromQuery{}))
		require.NoError(t, err)

		assert.Equal(t, "left", qi.Tables[1].JoinType)
	})

	t.Run("select and sort indices", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQuery{}))
		require.NoError(t, err)

		// All fields are selectable
		assert.Equal(t, 5, len(qi.SelectIdxList))

		// Only user_row.name has sort=1
		assert.Equal(t, 1, len(qi.SortIdxList))
		assert.Equal(t, "user_row.name", qi.CombinedFields[qi.SortIdxList[0]].SQLName)
	})

	t.Run("field name index", func(t *testing.T) {
		qi, err := struct_info.CollectQueryInfo(reflect.TypeOf(UserCommentsQuery{}))
		require.NoError(t, err)

		assert.Equal(t, 0, qi.FieldNameIdx["user_row.id"])
		assert.Equal(t, 1, qi.FieldNameIdx["user_row.name"])
		assert.Equal(t, 2, qi.FieldNameIdx["c1.id"])
		assert.Equal(t, 3, qi.FieldNameIdx["c1.user_id"])
		assert.Equal(t, 4, qi.FieldNameIdx["c1.text"])
	})

	t.Run("error: no fields", func(t *testing.T) {
		type EmptyQuery struct{}

		_, err := struct_info.CollectQueryInfo(reflect.TypeOf(EmptyQuery{}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no non-omit fields")
	})

	t.Run("error: multiple FROM", func(t *testing.T) {
		type MultiFromQuery struct {
			First  UserRow `tbl:"from"`
			Second UserRow `tbl:"from"`
		}

		_, err := struct_info.CollectQueryInfo(reflect.TypeOf(MultiFromQuery{}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple FROM")
	})

	t.Run("error: multiple PK", func(t *testing.T) {
		type MultiPKQuery struct {
			First  UserRow `tbl:"pk"`
			Second UserRow `tbl:"pk"`
		}

		_, err := struct_info.CollectQueryInfo(reflect.TypeOf(MultiPKQuery{}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple PK")
	})
}
