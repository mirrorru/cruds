# Структура проекта Quick CRUD

## Обзор файловой структуры

### Корневой пакет (cruds)
- `table.go` — универсальная реализация `Table[ROW]` с использованием reflection
- `filter.go` — дерево условий фильтрации (ConditionNode, GroupNode, FilterNode)
- `filter_test.go` — тесты системы фильтрации
- `contracts.go` — интерфейсы и контракты (`TypedTable[ROW]`, `TxProcessor`, `Row`, `Result`, `Rows`)
- `dialect_vars.go` — пакетные переменные-алиасы для SQL-диалектов (`SQLite`, `PostgresSQL`), предоставляющие удобный доступ к реализациям диалектов без импорта пакета `dialect`

### cmd/qcgen
Генератор кода для создания типизированных реализаций таблиц без reflection.
- `main.go` — генератор, который парсит AST Go файлов и создаёт hard-coded реализации

**Основные возможности генератора:**
- Парсинг полей структуры из AST
- Обработка embedded структур
- Извлечение SQL имени из метода `SQLName()`
- Генерация hard-coded `TableInfo` без reflection
- Генерация helper методов: `scanRefs`, `insertArgs`, `updateArgs`
- Генерация CRUD методов с прямым доступом к полям

**Использование:**
```bash
go install ./cmd/qcgen
qcgen -src=path/to/models:*Row -dest=path/to/repo -pkg=repo
```

### dialect
SQL диалекты:
- `dialects.go` — общий интерфейс диалекта и вспомогательные функции
- `postgres.go` — PostgreSQL диалект
- `sqlite.go` — SQLite диалект
- `dialect_test.go` — тесты диалектов

### struct_info
Метаданные таблиц и полей:
- `field_tags.go` — парсинг тегов `tbl`
- `table_field.go` — `TableField`, `TableFields`, `ExtractArgs`, `ExtractRefs`
- `table_info.go` — `TableInfo`, `GetTableInfo`
- `table_info_test.go` — тесты
- `table_sql_texts.go` — построение SQL текстов

### tx_adapter
Адаптеры для работы с БД:
- `pgx_adapter.go` — адаптеры для pgx.Conn и pgx.Tx
- `sql_adapter.go` — адаптеры для *sql.DB и *sql.Tx

### test
Тестовые файлы:
- `gen/model/models.go` — тестовые модели (UserRow, ProductRow)
- `gen/repo/` — сгенерированные реализации таблиц
- `gensrc/` — исходные файлы для генерации
- `smoke/` — smoke тесты
- `samples/` — примеры структур

## Примечания по линтеру

Оставшиеся предупреждения линтера (не критично):
- `gocognit` — высокая когнитивная сложность функций генератора
- `lll` — длинные строки в шаблоне (ожидаемо для генерации кода)
- `nestif` — сложные вложенные блоки
- `staticcheck` — устаревший `parser.ParseDir` (предупреждение о депрекации)

Все критические ошибки линтера исправлены.
