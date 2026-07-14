package dialect

type SQLDialect interface {
	Name() string                               // Название диалекта. / EN: Dialect name.
	Placeholder(pos int) string                 // Плейсхолдер для параметра (1-based). / EN: Placeholder for parameter (1-based).
	SupportsReturning() bool                    // Поддерживает ли RETURNING. / EN: Whether RETURNING is supported.
	OffsetAndLimit(offset, limit uint32) string // Строка с Offset и Limit
	ILIKE(col string, placeholder string) string
}
