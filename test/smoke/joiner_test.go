//go:build smoke

package smoke

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/tx_adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type JoinRowFrom struct {
	ID       int64 `tbl:"pk;auto"`
	Name     string
	Birthday ClientBirthday
	Gender   GenderType
}

func (JoinRowFrom) SQLName() string {
	return "table_from"
}

type JoinRowInnerJoin struct {
	ID        int64 `tbl:"pk;auto"`
	RefID     int64 `tbl:"ref=table_from:id"`
	InnerName string
}

func (JoinRowInnerJoin) SQLName() string {
	return "table_inner"
}

type JoinRowLeftJoin struct {
	ID       int64  `tbl:"pk;auto"`
	RefID    int64  `tbl:"ref=table_from:id"`
	LeftName string `tbl:"sort=1"`
	Birthday ClientBirthday
	Gender   GenderType
}

func (JoinRowLeftJoin) SQLName() string {
	return "table_left"
}

type JoinRowAnonymousVal struct {
	InnerVal JoinRowInnerJoin
	LeftVal  *JoinRowLeftJoin `tbl:"join=left;alias=LV;sort=10"`
}

type JoinSummary struct {
	From JoinRowFrom
	JoinRowAnonymousVal
}

func TestNewJoiner(t *testing.T) {
	t.Parallel()
	join, err := cruds.NewJoiner[JoinSummary](cruds.SQLite)
	require.NoError(t, err)
	require.NotNil(t, join)
	for idx, table := range join.Tables() {
		fmt.Printf("%d:\n%#v\n", idx, table)
	}
	fmt.Println("GetOne:", join.SQLs().GetOneSQL)
	fmt.Println("Sort:", join.SQLs().SortSQL)
}

func TestJoinerOne(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	txAdapter := tx_adapter.NewDBAdapterVal(db)
	{
		_, err = db.ExecContext(ctx, `
		CREATE TABLE table_from (
			id integer primary key autoincrement not null,
			name text not null,
		    birthday date not null,
		    gender text not null
		)
	`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `
		CREATE TABLE table_inner (
			id integer primary key autoincrement not null,
			ref_id int not null,
			inner_name text not null
		)
	`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `
		CREATE TABLE table_left (
			id integer primary key autoincrement not null,
			ref_id int not null,
			left_name text not null,
		    birthday date not null,
		    gender text not null
		)
	`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `
		INSERT INTO table_from (id, name, birthday, gender) VALUES (1, 'from_1', '2001-02-03', 'male'), (2, 'from_2', '2004-05-06', 'female')
	`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `
		INSERT INTO table_inner (id, ref_id, inner_name) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')
	`)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `
		INSERT INTO table_left (id, ref_id, left_name, birthday, gender) VALUES (100, 1, 'left_1', '2003-02-01', 'male')
	`)
		require.NoError(t, err)
	}
	join, err := cruds.NewJoiner[JoinSummary](cruds.SQLite)
	require.NoError(t, err)

	result, err := join.One(ctx, txAdapter, 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(1), result.From.ID)
	assert.Equal(t, "from_1", result.From.Name)
	assert.Equal(t, "2001-02-03 00:00:00 +0000 UTC", result.From.Birthday.String())
	assert.Equal(t, GenderTypeMale, result.From.Gender)

	require.NotNil(t, result.InnerVal)
	assert.Equal(t, "inner_1", result.InnerVal.InnerName)
	require.NotNil(t, result.LeftVal)
	assert.Equal(t, "left_1", result.LeftVal.LeftName)

	result2, err := join.One(ctx, txAdapter, 2)
	require.NoError(t, err)
	require.NotNil(t, result2)
	require.Equal(t, int64(2), result2.From.ID)
	assert.Equal(t, "2004-05-06 00:00:00 +0000 UTC", result2.From.Birthday.String())
	assert.Equal(t, GenderTypeFemale, result2.From.Gender)
	assert.Equal(t, "from_2", result2.From.Name)
	require.NotNil(t, result2.InnerVal)
	assert.Equal(t, "inner_2", result2.InnerVal.InnerName)
	require.Nil(t, result2.LeftVal)

	result3, err := join.One(ctx, txAdapter, 999)
	require.Error(t, err)
	_ = result3
}
