package quick_crud

import "quick-crud/dialect"

var (
	// SQLite реализует диалект SQLite
	SQLite dialect.SQLiteDialect // реализует диалект SQLite
	_      = SQLite

	// PostgresSQL реализует диалект PostgresSQL
	PostgresSQL dialect.PostgreSQLDialect
	_           = PostgresSQL
)
