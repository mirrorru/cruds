//go:build smoke

package smoke

import (
	"testing"

	"github.com/mirrorru/crudquick/tx_adapter"

	qc "github.com/mirrorru/crudquick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const querySetupSQL = `
	DROP TABLE IF EXISTS comment_row;
	DROP TABLE IF EXISTS post_row;
	DROP TABLE IF EXISTS user_row;

	CREATE TABLE user_row (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER
	);
	CREATE TABLE comment_row (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		text TEXT,
		FOREIGN KEY (user_id) REFERENCES user_row(id)
	);
	CREATE TABLE post_row (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT,
		FOREIGN KEY (user_id) REFERENCES user_row(id)
	);
`

func setupQueryTables(t *testing.T, env *testEnv) {
	t.Helper()
	_, err := env.db.Exec(querySetupSQL)
	require.NoError(t, err)
}

func TestQueryOne(t *testing.T) {
	env := newTestEnv(t)
	tx := tx_adapter.NewDBAdapterVal(env.db)
	setupQueryTables(t, env)

	_, err := env.db.Exec(`
		INSERT INTO user_row (id, name, age) VALUES (1, 'Alice', 30);
		INSERT INTO comment_row (user_id, text) VALUES (1, 'Hello');
		INSERT INTO comment_row (user_id, text) VALUES (1, 'World');
	`)
	require.NoError(t, err)

	t.Run("One returns single record", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 1, result.User.ID)
		assert.Equal(t, "Alice", result.User.Name)
		assert.Equal(t, 30, result.User.Age)
		assert.Equal(t, 1, result.Comment.UserID)
	})

	t.Run("One with pointer field", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQueryPtr](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 1, result.User.ID)
		assert.NotNil(t, result.Comment)
		assert.Equal(t, 1, result.Comment.UserID)
	})

	t.Run("One not found", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)
		result, err := q.One(env.ctx, tx, 999)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestQueryMany(t *testing.T) {
	env := newTestEnv(t)
	tx := tx_adapter.NewDBAdapterVal(env.db)
	setupQueryTables(t, env)

	_, err := env.db.Exec(`
		INSERT INTO user_row (id, name, age) VALUES (1, 'Alice', 30);
		INSERT INTO user_row (id, name, age) VALUES (2, 'Bob', 25);
		INSERT INTO comment_row (user_id, text) VALUES (1, 'Comment 1');
		INSERT INTO comment_row (user_id, text) VALUES (1, 'Comment 2');
		INSERT INTO comment_row (user_id, text) VALUES (2, 'Comment 3');
	`)
	require.NoError(t, err)

	t.Run("Many returns all records", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)
		results, err := q.Many(env.ctx, tx, nil)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 3, len(results))
	})

	t.Run("Many with filter", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)

		// CombinedFields: [0]=user_row.id, [1]=user_row.name, [2]=user_row.age, [3]=c1.id, [4]=c1.user_id, [5]=c1.text
		f := &qc.Filter{
			Range: qc.And(
				qc.Cond(1, qc.CmdEq, "Alice"),
			),
		}

		results, err := q.Many(env.ctx, tx, f)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 2, len(results))
		for _, r := range results {
			assert.Equal(t, "Alice", r.User.Name)
		}
	})

	t.Run("Many with limit", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)

		f := &qc.Filter{
			Limit: 2,
		}

		results, err := q.Many(env.ctx, tx, f)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 2, len(results))
	})

	t.Run("Many with offset", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)

		f := &qc.Filter{
			Offset: 1,
			Limit:  2,
		}

		results, err := q.Many(env.ctx, tx, f)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 2, len(results))
	})
}

func TestQueryNULLHandling(t *testing.T) {
	env := newTestEnv(t)
	tx := tx_adapter.NewDBAdapterVal(env.db)
	setupQueryTables(t, env)

	_, err := env.db.Exec(`
		INSERT INTO user_row (id, name, age) VALUES (1, 'Charlie', 35);
	`)
	require.NoError(t, err)

	t.Run("LEFT JOIN with NULL pointer field", func(t *testing.T) {
		q := qc.NewQueryVal[UserCommentsQueryPtr](qc.SQLite)
		results, err := q.Many(env.ctx, tx, nil)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Charlie", results[0].User.Name)
		assert.Nil(t, results[0].Comment)
	})
}

func TestQueryInternals(t *testing.T) {
	q := qc.NewQueryVal[UserCommentsQuery](qc.SQLite)
	internals := q.Internals()

	assert.NotNil(t, internals.QueryInfo)
	assert.NotNil(t, internals.SqlTexts)
	assert.NotNil(t, internals.CombinedFields)

	assert.Equal(t, 2, len(internals.QueryInfo.Tables))
	assert.Equal(t, 6, len(internals.CombinedFields))
}

const customTypeSetupSQL = `
	DROP TABLE IF EXISTS orders;
	DROP TABLE IF EXISTS clients;

	CREATE TABLE clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		birthday TEXT,
		gender TEXT
	);

	CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id INTEGER,
		item TEXT,
		qty INTEGER,
		FOREIGN KEY (client_id) REFERENCES clients(id)
	);
`

func setupCustomTypeTables(t *testing.T, env *testEnv) {
	t.Helper()
	_, err := env.db.Exec(customTypeSetupSQL)
	require.NoError(t, err)
}

func TestQueryWithCustomTypes(t *testing.T) {
	env := newTestEnv(t)
	tx := tx_adapter.NewDBAdapterVal(env.db)
	setupCustomTypeTables(t, env)

	_, err := env.db.Exec(`
		INSERT INTO clients (id, name, birthday, gender) VALUES 
			(1, 'Иван Иванов', '1990-05-15', 'male'),
			(2, 'Мария Петрова', '1985-12-03', 'female'),
			(3, 'Алексей Сидоров', '1995-07-20', 'male');

		INSERT INTO orders (client_id, item, qty) VALUES 
			(1, 'Ноутбук', 1),
			(1, 'Мышь', 2),
			(2, 'Клавиатура', 1);
	`)
	require.NoError(t, err)

	t.Run("One with non-pointer fields", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQuery](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, int64(1), result.Client.ID)
		assert.Equal(t, ClientName("Иван Иванов"), result.Client.Name)
		assert.Equal(t, GenderTypeMale, result.Client.Gender)
		assert.Equal(t, int64(1), result.Order.ID)
		assert.Equal(t, "Ноутбук", result.Order.Item)
	})

	t.Run("One with pointer Client", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQueryPtrClient](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Client)

		assert.Equal(t, int64(1), result.Client.ID)
		assert.Equal(t, ClientName("Иван Иванов"), result.Client.Name)
		assert.Equal(t, GenderTypeMale, result.Client.Gender)
	})

	t.Run("One with pointer Order", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQueryPtrOrder](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Order)

		assert.Equal(t, int64(1), result.Client.ID)
		assert.Equal(t, int64(1), result.Order.ID)
		assert.Equal(t, "Ноутбук", result.Order.Item)
	})

	t.Run("One with both pointer fields", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQueryPtrBoth](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Client)
		require.NotNil(t, result.Order)

		assert.Equal(t, int64(1), result.Client.ID)
		assert.Equal(t, int64(1), result.Order.ID)
	})

	t.Run("Many returns all records", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQuery](qc.SQLite)
		results, err := q.Many(env.ctx, tx, nil)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 4, len(results))
	})

	t.Run("Many with filter", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQuery](qc.SQLite)

		f := &qc.Filter{
			Range: qc.And(
				qc.Cond(1, qc.CmdEq, "Иван Иванов"),
			),
		}

		results, err := q.Many(env.ctx, tx, f)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 2, len(results))
		for _, r := range results {
			assert.Equal(t, ClientName("Иван Иванов"), r.Client.Name)
		}
	})

	t.Run("LEFT JOIN with NULL pointer Order", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQueryPtrOrder](qc.SQLite)

		f := &qc.Filter{
			Range: qc.And(
				qc.Cond(0, qc.CmdEq, int64(3)),
			),
		}

		results, err := q.Many(env.ctx, tx, f)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 1, len(results))
		assert.Equal(t, ClientName("Алексей Сидоров"), results[0].Client.Name)
		assert.Nil(t, results[0].Order)
	})

	t.Run("LEFT JOIN with NULL both pointer fields", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQueryPtrBoth](qc.SQLite)

		f := &qc.Filter{
			Range: qc.And(
				qc.Cond(0, qc.CmdEq, int64(3)),
			),
		}

		results, err := q.Many(env.ctx, tx, f)
		require.NoError(t, err)
		require.NotNil(t, results)

		assert.Equal(t, 1, len(results))
		require.NotNil(t, results[0].Client)
		assert.Equal(t, ClientName("Алексей Сидоров"), results[0].Client.Name)
		assert.Nil(t, results[0].Order)
	})

	t.Run("Custom type scanning", func(t *testing.T) {
		q := qc.NewQueryVal[ClientOrdersQuery](qc.SQLite)
		result, err := q.One(env.ctx, tx, 1)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, ClientName("Иван Иванов"), result.Client.Name)
		assert.Equal(t, GenderTypeMale, result.Client.Gender)
		assert.NotZero(t, result.Client.Birthday)
	})
}
