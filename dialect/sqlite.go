package dialect

import "fmt"

// SQLiteDialect реализует диалект SQLite.
// EN: SQLiteDialect implements the SQLite dialect.
type SQLiteDialect struct{}

// Проверка интерфейса на этапе компиляции.
// EN: Interface check at compile time.
var _ SQLDialect = SQLiteDialect{}

// Name возвращает название диалекта.
// EN: Name returns the dialect name.
func (SQLiteDialect) Name() string { return "sqlite" }

func (SQLiteDialect) Placeholder(_ int) string { return "?" }

func (SQLiteDialect) SupportsReturning() bool { return true }
func (SQLiteDialect) ILIKE(col string, placeholder string) string {
	return "LOWER(" + col + ") LIKE LOWER(" + placeholder + ")"
}

func (SQLiteDialect) OffsetAndLimit(offset, limit uint32) string {
	if limit == 0 && offset == 0 {
		return ""
	}
	if offset == 0 {
		return fmt.Sprintf(" LIMIT %d", limit)
	}
	if limit == 0 {
		return fmt.Sprintf(" LIMIT -1 OFFSET %d", offset)
	}
	return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}
