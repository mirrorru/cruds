//go:build functional

package functional

import (
	"fmt"
	"testing"

	"github.com/mirrorru/cruds/dialect"
	"github.com/mirrorru/cruds/test/test_run"
)

func TestUserRow_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, age INT NOT NULL)", test_run.UserRowCRUDTable))
	test_run.UserRow_Reflection_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestUserRow_Typed_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, age INT NOT NULL)", test_run.UserRowTypedCRUDTable))
	test_run.UserRow_Typed_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestProductRow_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, price REAL NOT NULL, stock INT NOT NULL)", test_run.ProductRowCRUDTable))
	test_run.ProductRow_Reflection_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestProductRow_Typed_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, price REAL NOT NULL, stock INT NOT NULL)", test_run.ProductRowTypedCRUDTable))
	test_run.ProductRow_Typed_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestFuncRow_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id SERIAL PRIMARY KEY, name TEXT NOT NULL, value INT NOT NULL)", test_run.FuncRowCRUDTable))
	test_run.FuncRow_Reflection_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestFuncRow_Typed_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id SERIAL PRIMARY KEY, name TEXT NOT NULL, value INT NOT NULL)", test_run.FuncRowTypedCRUDTable))
	test_run.FuncRow_Typed_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestIdNameAgeRowFilled_Reflection_CRUD(t *testing.T) {
	t.Parallel()
	sharedExec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id SERIAL PRIMARY KEY, name TEXT NOT NULL, age INT NULL)", test_run.IdNameAgeRowFilledCRUDTable))
	test_run.IdNameAgeRowFilled_Reflection_CRUD(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestJoinSample_Reflection_OneMany(t *testing.T) {
	sharedExec("CREATE TABLE IF NOT EXISTS join_from (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS join_inner (id BIGSERIAL PRIMARY KEY, ref_id BIGINT NOT NULL, inner_val TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS join_left (id BIGSERIAL PRIMARY KEY, ref_id BIGINT NOT NULL, left_val TEXT NOT NULL)")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1')")
	test_run.JoinSample_Reflection_OneMany(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestJoinSample_Reflection_NullPointer(t *testing.T) {
	sharedExec("DELETE FROM join_left")
	sharedExec("DELETE FROM join_inner")
	sharedExec("DELETE FROM join_from")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2'), (3, 'from_3')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2'), (30, 3, 'inner_3')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1'), (300, 3, 'left_3')")
	test_run.JoinSample_Reflection_NullPointer(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestJoinSample_Typed_OneMany(t *testing.T) {
	sharedExec("DELETE FROM join_left")
	sharedExec("DELETE FROM join_inner")
	sharedExec("DELETE FROM join_from")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1')")
	test_run.JoinSample_Typed_OneMany(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestJoinSample_Typed_NullPointer(t *testing.T) {
	sharedExec("DELETE FROM join_left")
	sharedExec("DELETE FROM join_inner")
	sharedExec("DELETE FROM join_from")
	sharedExec("INSERT INTO join_from (id, name) VALUES (1, 'from_1'), (2, 'from_2'), (3, 'from_3')")
	sharedExec("INSERT INTO join_inner (id, ref_id, inner_val) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2'), (30, 3, 'inner_3')")
	sharedExec("INSERT INTO join_left (id, ref_id, left_val) VALUES (100, 1, 'left_1'), (300, 3, 'left_3')")
	test_run.JoinSample_Typed_NullPointer(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestJoinSummary_Reflection_OneMany(t *testing.T) {
	sharedExec("CREATE TABLE IF NOT EXISTS table_from (id BIGSERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL, birthday DATE NOT NULL, gender TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS table_inner (id BIGSERIAL PRIMARY KEY NOT NULL, ref_id INT NOT NULL, inner_name TEXT NOT NULL)")
	sharedExec("CREATE TABLE IF NOT EXISTS table_left (id BIGSERIAL PRIMARY KEY NOT NULL, ref_id INT NOT NULL, left_name TEXT NOT NULL, birthday DATE NOT NULL, gender TEXT NOT NULL)")
	sharedExec("DELETE FROM table_left")
	sharedExec("DELETE FROM table_inner")
	sharedExec("DELETE FROM table_from")
	sharedExec("INSERT INTO table_from (id, name, birthday, gender) VALUES (1, 'from_1', '2001-02-03', 'male'), (2, 'from_2', '2004-05-06', 'female')")
	sharedExec("INSERT INTO table_inner (id, ref_id, inner_name) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO table_left (id, ref_id, left_name, birthday, gender) VALUES (100, 1, 'left_1', '2003-02-01', 'male')")
	test_run.JoinSummary_Reflection_OneMany(t, sharedTx(), dialect.PostgreSQLDialect{})
}

func TestJoinSummary_Typed_OneMany(t *testing.T) {
	sharedExec("DELETE FROM table_left")
	sharedExec("DELETE FROM table_inner")
	sharedExec("DELETE FROM table_from")
	sharedExec("INSERT INTO table_from (id, name, birthday, gender) VALUES (1, 'from_1', '2001-02-03', 'male'), (2, 'from_2', '2004-05-06', 'female')")
	sharedExec("INSERT INTO table_inner (id, ref_id, inner_name) VALUES (10, 1, 'inner_1'), (20, 2, 'inner_2')")
	sharedExec("INSERT INTO table_left (id, ref_id, left_name, birthday, gender) VALUES (100, 1, 'left_1', '2003-02-01', 'male')")
	test_run.JoinSummary_Typed_OneMany(t, sharedTx(), dialect.PostgreSQLDialect{})
}
