package quick_crud

import "quick-crud/dialect"

var (
	// SQLite алиас для реализации диалекта SQLite
	SQLite dialect.SQLiteDialect
	_      = SQLite

	// PostgresSQL алиас для реализации диалекта PostgresSQL
	PostgresSQL dialect.PostgreSQLDialect
	_           = PostgresSQL
)
