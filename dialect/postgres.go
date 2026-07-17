package dialect

import (
	"fmt"
	"strconv"

	"github.com/mirrorru/cruds/defs"
)

type PostgreSQLDialect struct{}

var _ SQLDialect = PostgreSQLDialect{}

func (PostgreSQLDialect) Name() string { return "postgres" }

func (PostgreSQLDialect) Placeholder(pos int) string { return "$" + strconv.Itoa(pos) }

func (PostgreSQLDialect) SupportsReturning() bool { return true }

func (PostgreSQLDialect) ILIKE(col string, placeholder string) string {
	return col + defs.SQLILike + placeholder
}

func (PostgreSQLDialect) OffsetAndLimit(offset, limit uint32) string {
	if limit == 0 && offset == 0 {
		return ""
	}
	return fmt.Sprintf(" OFFSET %d LIMIT %d", offset, limit)
}
