# Структура проекта CRUDs

## Обзор файловой структуры

### Корневой пакет (cruds)
- `table.go` — универсальная реализация `Table[ROW]` с использованием reflection
- `joiner.go` — реализация `Joiner[JT]` для SELECT с JOIN нескольких таблиц
- `jointable_tag_flags.go` — парсинг тегов полей join-структуры (`ParseJoinTableFlags`)
- `filter.go` — дерево условий фильтрации (ConditionNode, GroupNode, FilterNode)
- `filter_test.go` — тесты системы фильтрации
- `contracts.go` — интерфейсы и контракты (`TypedTable[ROW]`, `TypedJoiner[JT]`, `TxProcessor`, `Row`, `Result`, `Rows`)
- `dialect_vars.go` — пакетные переменные-алиасы для SQL-диалектов (`SQLite`, `PostgresSQL`), предоставляющие удобный доступ к реализациям диалектов без импорта пакета `dialect`

### cmd/crudsgen
Генератор кода для создания типизированных реализаций таблиц и join-структур без reflection.
- `main.go` — генератор, который парсит AST Go файлов и создаёт hard-coded реализации

**Основные возможности генератора:**
- Парсинг полей структуры из AST
- Обработка embedded структур
- Извлечение SQL имени из метода `SQLName()`
- Генерация hard-coded `TableInfo` без reflection
- Генерация helper методов: `scanRefs`, `insertArgs`, `updateArgs`
- Генерация CRUD методов с прямым доступом к полям
- **Генерация Joiner* структур** — типизированные реализации join-запросов (INNER/LEFT/RIGHT/OUTER/CROSS JOIN) с pre-computed SQL и hardcoded `makeRefs`/`applyRefs`
- **Поддержка build-тегов** через флаг `-build=` (можно указывать несколько раз, объединяются через `||`)

**Использование:**
```bash
go install ./cmd/crudsgen
# Генерация Table* (по умолчанию)
crudsgen -src=path/to/models:*Row -dest=path/to/repo -pkg=repo
# Генерация Joiner*
crudsgen -src=path/to/models:Join* -dest=path/to/repo -pkg=repo -joiner
# Генерация с build-тегами
crudsgen -src=path/to/models:*Row -dest=path/to/repo -pkg=repo -build=crudsgen
```

**Флаги:**
- `-src` — спецификация исходных файлов (path:pattern), можно указывать несколько раз
- `-dest` — директория для генерации (обязательный)
- `-pkg` — имя пакета (по умолчанию: имя директории dest)
- `-no-genstr` — не добавлять `//go:generate` комментарий
- `-build` — build-тег для добавления в генерируемый файл (можно указывать несколько раз)
- `-table` — генерировать Table* (по умолчанию true, если не указан `-joiner`)
- `-joiner` — генерировать Joiner* для join-структур
- `-help` — показать справку

### dialect
SQL диалекты:
- `dialects.go` — общий интерфейс диалекта и вспомогательные функции
- `postgres.go` — PostgreSQL диалект
- `sqlite.go` — SQLite диалект
- `dialect_test.go` — тесты диалектов

### struct_info
Метаданные таблиц и полей:
- `field_tags.go` — парсинг тегов `crud`
- `table_field.go` — `TableField`, `TableFields`, `ExtractArgs`, `ExtractRefs`
- `table_info.go` — `TableInfo`, `GetTableInfo`
- `table_info_test.go` — тесты
- `table_sql_texts.go` — построение SQL текстов

### tx_adapter
Адаптеры для работы с БД:
- `pgx_adapter.go` — адаптеры для pgx.Conn и pgx.Tx
- `sql_adapter.go` — адаптеры для *sql.DB и *sql.Tx
- `pgxpool_adapter.go` — адаптер для *pgxpool.Pool (пул соединений PostgreSQL)

### test
Тестовые файлы:
- `testmodels/` — единый пакет моделей (`package testmodels`, build-теги: `smoke || crudsgen || functional`):
  - `models.go` — `UserRow`, `ProductRow`, `IdNameAgeRowFilled`, `FuncRow`
  - `join_models.go` — `JoinFromRow`, `JoinInnerRow`, `JoinLeftRow`, `JoinSample`
  - `join_summary.go` — `JoinRowFrom`, `JoinRowInnerJoin`, `JoinRowLeftJoin`, `JoinRowAnonymousVal`, `JoinSummary`
  - `types.go` — `ClientName`, `ClientBirthday`, `Date`
  - `enum.go` — `GenderType`, `makeScanMap`, `scan`
  - Сгенерированные: `table_*.go` (Table-реализации с опциональным `tableName ...string`), `joiner_*.go` (Joiner-реализации с опциональным `suffix ...string`)
- `test_run/` — общая логика тестов (`package test_run`, build-теги: `smoke || functional`):
  - `table_userrow.go`, `table_productrow.go`, `table_funcrow.go`, `table_idnameage.go` — CRUD-тесты (reflection через marker-структуры с переопределённым `SQLName()` + typed через опциональный `tableName`)
  - Константы имён таблиц: `UserRowCRUDTable`, `UserRowTypedCRUDTable`, `ProductRowCRUDTable`, `ProductRowTypedCRUDTable`, `FuncRowCRUDTable`, `FuncRowTypedCRUDTable`, `IdNameAgeRowFilledCRUDTable`
  - Marker-типы: `userRowCRUD`, `productRowCRUD`, `funcRowCRUD`, `idNameAgeRowFilledCRUD` — обёртки моделей с переопределённым `SQLName()`
  - `joiner_joinsample.go`, `joiner_joinsummary.go` — Joiner-тесты (reflection на оригинальных типах — последовательно, typed с суффиксами — параллельно)
  - Константы суффиксов: `JoinSampleOneManySfx`, `JoinSampleNullPtrSfx`, `JoinSampleTypedOneManySfx`, `JoinSampleTypedNullPtrSfx`, `JoinSummaryOneManySfx`, `JoinSummaryTypedOneManySfx`
- `smoke/` — smoke-тесты на SQLite (build tag: `smoke`):
  - `main_test.go` — глобальный `sharedDB` (`file::memory:?cache=shared`), `TestMain`, `sharedTx()`/`sharedExec()`
  - `smoke_test.go` — все обёртки тестов (CRUD, join, corner) с `t.Parallel()` (кроме reflection joiner-тестов — последовательные с общими таблицами)
- `crudsgen/` — проверки интерфейсов сгенерированного кода (build tag: `crudsgen`):
  - `compile_test.go` — `var _ cruds.TypedTable[...] = (*...)(nil)` и аналогичные проверки контрактов
- `samples/` — примеры структур для unit-тестов `struct_info` (без build-тегов)
- `functional/` — functional-тесты на PostgreSQL (build tag: `functional`):
  - `main_test.go` — глобальный `sharedPool` (`*pgxpool.Pool`), `TestMain`, `sharedTx()`/`sharedExec()`
  - `functional_test.go` — обёртки вокруг `test_run` функций с PostgreSQL-синтаксисом DDL, включая JoinSummary тесты

## Стратегия изоляции тестов

- **Smoke (SQLite)**: одна общая in-memory БД (`file::memory:?cache=shared`), `SetMaxOpenConns(1)`, все тесты разделяют одно подключение
- **Functional (PostgreSQL)**: общий пул соединений `*pgxpool.Pool` (через `pgxpool.New`)
- **Изоляция таблиц**: CRUD-тесты и typed-joiner-тесты используют уникальные имена таблиц (константы/суффиксы из `test_run`) и запускаются параллельно (`t.Parallel()`)
- **Reflection-joiner-тесты**: используют оригинальные имена таблиц (теги `ref=` не наследуются через embedding marker-типов), запускаются последовательно с очисткой данных между тестами
- **Инициализация**: `TestMain` в `smoke/main_test.go` и `functional/main_test.go`

## Примечания по линтеру

Оставшиеся предупреждения линтера (не критично):
- `gocognit` — высокая когнитивная сложность функций генератора
- `lll` — длинные строки в шаблоне (ожидаемо для генерации кода)
- `nestif` — сложные вложенные блоки
- `staticcheck` — устаревший `parser.ParseDir` (предупреждение о депрекации)

Все критические ошибки линтера исправлены.
