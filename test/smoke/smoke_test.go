//go:build smoke

package smoke

import (
	"fmt"
	"context"
	"sync"
	"testing"

	"github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/dialect"
	"github.com/mirrorru/cruds/test/test_run"
	"github.com/mirrorru/cruds/test/testmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRow_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, age INTEGER NOT NULL)", test_run.UserRowCRUDTable))
	test_run.UserRow_Reflection_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestUserRow_Typed_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, age INTEGER NOT NULL)", test_run.UserRowTypedCRUDTable))
	test_run.UserRow_Typed_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestProductRow_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL, stock INTEGER NOT NULL)", test_run.ProductRowCRUDTable))
	test_run.ProductRow_Reflection_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestProductRow_Typed_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL, stock INTEGER NOT NULL)", test_run.ProductRowTypedCRUDTable))
	test_run.ProductRow_Typed_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestFuncRow_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, \"value\" INTEGER NOT NULL)", test_run.FuncRowCRUDTable))
	test_run.FuncRow_Reflection_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestFuncRow_Typed_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, \"value\" INTEGER NOT NULL)", test_run.FuncRowTypedCRUDTable))
	test_run.FuncRow_Typed_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestIdNameAgeRowFilled_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, age INT NULL)", test_run.IdNameAgeRowFilledCRUDTable))
	test_run.IdNameAgeRowFilled_Reflection_CRUD(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestNewJoiner(t *testing.T) {
	join, err := cruds.NewJoiner[testmodels.JoinSummary](dialect.SQLiteDialect{})
	require.NoError(t, err)
	require.NotNil(t, join)
	for idx, table := range join.Tables() {
		fmt.Printf("%d:\n%#v\n", idx, table)
	}
	fmt.Println("GetOne:", join.SQLs().GetOneSQL)
	fmt.Println("Sort:", join.SQLs().SortSQL)
}

func TestJoinSample_Reflection_OneMany(t *testing.T) {
	sharedExec("CREATE TABLE IF NOT EXISTS join_from (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS join_inner (id INTEGER PRIMARY KEY AUTOINCREMENT, ref_id INTEGER NOT NULL, inner_val TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS join_left (id INTEGER PRIMARY KEY AUTOINCREMENT, ref_id INTEGER NOT NULL, left_val TEXT NOT NULL)")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1')")
	test_run.JoinSample_Reflection_OneMany(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestJoinSample_Reflection_NullPointer(t *testing.T) {
	sharedExec("DELETE FROM join_left")
	sharedExec("DELETE FROM join_inner")
	sharedExec("DELETE FROM join_from")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2'), (3, 'from_3')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2'), (30, 3, 'inner_3')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1'), (300, 3, 'left_3')")
	test_run.JoinSample_Reflection_NullPointer(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestJoinSample_Typed_OneMany(t *testing.T) {
	sharedExec("DELETE FROM join_left")
	sharedExec("DELETE FROM join_inner")
	sharedExec("DELETE FROM join_from")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1')")
	test_run.JoinSample_Typed_OneMany(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestJoinSample_Typed_NullPointer(t *testing.T) {
	sharedExec("DELETE FROM join_left")
	sharedExec("DELETE FROM join_inner")
	sharedExec("DELETE FROM join_from")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2'), (3, 'from_3')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2'), (30, 3, 'inner_3')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1'), (300, 3, 'left_3')")
	test_run.JoinSample_Typed_NullPointer(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestJoinSummary_Reflection_OneMany(t *testing.T) {
	sharedExec("CREATE TABLE IF NOT EXISTS table_from (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name TEXT NOT NULL, birthday DATE NOT NULL, gender TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS table_inner (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, ref_id INT NOT NULL, inner_name TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS table_left (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, ref_id INT NOT NULL, left_name TEXT NOT NULL, birthday DATE NOT NULL, gender TEXT NOT NULL)")
	sharedExec("DELETE FROM table_left")
	sharedExec("DELETE FROM table_inner")
	sharedExec("DELETE FROM table_from")
	sharedExec("INSERT INTO table_from (id, name, birthday, gender) VALUES (1, 'from_1', '2001-02-03', 'male'), (2, 'from_2', '2004-05-06', 'female')")
	sharedExec("INSERT INTO table_inner (id, ref_id, inner_name) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO table_left (id, ref_id, left_name, birthday, gender) VALUES (100, 1, 'left_1', '2003-02-01', 'male')")
	test_run.JoinSummary_Reflection_OneMany(t, sharedTx(), dialect.SQLiteDialect{})
}

func TestJoinSummary_Typed_OneMany(t *testing.T) {
	sharedExec("DELETE FROM table_left")
	sharedExec("DELETE FROM table_inner")
	sharedExec("DELETE FROM table_from")
	sharedExec("INSERT INTO table_from (id, name, birthday, gender) VALUES (1, 'from_1', '2001-02-03', 'male'), (2, 'from_2', '2004-05-06', 'female')")
	sharedExec("INSERT INTO table_inner (id, ref_id, inner_name) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO table_left (id, ref_id, left_name, birthday, gender) VALUES (100, 1, 'left_1', '2003-02-01', 'male')")
	test_run.JoinSummary_Typed_OneMany(t, sharedTx(), dialect.SQLiteDialect{})
}

const UserRowConcurrentTable = "user_row_concurrent"

type userRowConcurrent struct {
	testmodels.UserRow
}

func (userRowConcurrent) SQLName() string { return UserRowConcurrentTable }

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, age INTEGER NOT NULL)", UserRowConcurrentTable))

	d := dialect.SQLiteDialect{}
	table := cruds.NewTable[userRowConcurrent](d)

	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _, err := table.Ins(context.Background(), sharedTx(), &userRowConcurrent{UserRow: testmodels.UserRow{Name: "Concurrent", Age: idx}})
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("Concurrent insert error: %v", err)
	}
}

const UserRowCornerTable = "user_row_corner"

type userRowCorner struct {
	testmodels.UserRow
}

func (userRowCorner) SQLName() string { return UserRowCornerTable }

func TestTable_IdNameAgeRowRowCRUD(t *testing.T) {
	t.Parallel()

	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, age INT NULL)", UserRowCornerTable))
	tx := sharedTx()
	ctx := context.Background()

	table := cruds.NewTable[userRowCorner](dialect.SQLiteDialect{})

	colin, _, err := table.Ins(ctx, tx, &userRowCorner{UserRow: testmodels.UserRow{Name: "Colin", Age: 33}})
	require.NoError(t, err)

	bob, _, err := table.Ins(ctx, tx, &userRowCorner{UserRow: testmodels.UserRow{Name: "Bob", Age: 22}})
	require.NoError(t, err)

	alice, _, err := table.Ins(ctx, tx, &userRowCorner{UserRow: testmodels.UserRow{Name: "Alice", Age: 1}})
	require.NoError(t, err)
	require.Equal(t, &userRowCorner{UserRow: testmodels.UserRow{ID: 3, Name: "Alice", Age: 1}}, alice)

	alice2, err := table.One(ctx, tx, alice.ID)
	assert.NotSame(t, alice, alice2)
	assert.Equal(t, *alice, *alice2)

	alice2.Age = 11
	alice, _, err = table.Upd(ctx, tx, alice2)
	assert.NotSame(t, alice2, alice)
	assert.Equal(t, *alice2, *alice)

	rows, err := table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	require.Equal(t, []*userRowCorner{alice, bob, colin}, rows)

	_, err = table.Del(ctx, tx, bob.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	_, err = table.Del(ctx, tx, colin.ID)
	require.NoError(t, err)

	rows, err = table.Many(ctx, tx, nil)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}

func TestTable_IdNameAgeRowFilledStructInfo(t *testing.T) {
	t.Parallel()
	table := cruds.NewTable[testmodels.IdNameAgeRowFilled](dialect.SQLiteDialect{})
	tableInfo := table.Internals().TableInfo
	assert.Equal(t, "id_name_age_filled", tableInfo.SQLName)
	require.Len(t, tableInfo.Fields, 3)
	assert.Equal(t, "id", tableInfo.Fields[0].SQLName)
	assert.Equal(t, "name", tableInfo.Fields[1].SQLName)
	assert.Equal(t, "age", tableInfo.Fields[2].SQLName)
}
