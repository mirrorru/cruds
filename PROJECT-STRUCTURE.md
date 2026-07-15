# Структура проекта Quick CRUD

## Обзор файловой структуры

### Корневой пакет (quick-crud)
- `table.go` — универсальная реализация `Table[ROW]` с использованием reflection
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

### contracts
Интерфейсы и контракты:
- `TxProcessor` — интерфейс для работы с транзакциями
- `TypedTable[ROW]` — интерфейс для типизированных таблиц

### dialect
SQL диалекты:
- `postgres.go` — PostgreSQL диалект
- `sqlite.go` — SQLite диалект

### filter
Система фильтрации:
- `filter.go` — дерево условий (ConditionNode, GroupNode, FilterNode)

### helpers
Вспомогательные функции:
- `casing.go` — преобразование CamelCase в snake_case

### struct_info
Метаданные таблиц и полей:
- `field_tags.go` — парсинг тегов `tbl`
- `table_field.go` — `TableField`, `TableFields`, `ExtractArgs`, `ExtractRefs`
- `table_info.go` — `TableInfo`, `GetTableInfo`
- `table_sql_texts.go` — построение SQL текстов

### tx_adapter
Адаптеры для работы с БД:
- `pgx_adapter.go` — адаптеры для pgx.Conn и pgx.Tx
- `sql_adapter.go` — адаптеры для *sql.DB и *sql.Tx

### test
Тестовые файлы:
- `gen/model/models.go` — тестовые модели (UserRow, ProductRow)
- `gen/repo/` — сгенерированные реализации таблиц
- `smoke/` — smoke тесты
- `samples/` — примеры структур

## Примечания по линтеру

Оставшиеся предупреждения линтера (не критично):
- `gocognit` — высокая когнитивная сложность функций генератора
- `lll` — длинные строки в шаблоне (ожидаемо для генерации кода)
- `nestif` — сложные вложенные блоки
- `staticcheck` — устаревший `parser.ParseDir` (предупреждение о депрекации)

Все критические ошибки линтера исправлены.
