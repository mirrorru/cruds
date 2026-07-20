# Архитектура проекта CRUDs

## Подход

Domain Driven Design & Clean Architecture.

## Слои архитектуры

### Domain Layer (Доменный слой)

- **contracts.go (корневой пакет)** — интерфейсы и контракты (`TypedTable[ROW]`, `TypedJoiner[JT]`, `TxProcessor`, `Row`, `Result`, `Rows`)
- Определяют абстракции, не зависят от реализаций

### Application Layer (Слой приложения)

- **Корневой пакет (cruds)** — универсальная реализация `Table[ROW]`, `Joiner[JT]` для SELECT с JOIN нескольких таблиц, и система фильтрации (`Filter`, `FilterNode`, `ConditionNode`, `GroupNode`) с использованием reflection
- **jointable_tag_flags.go** — парсинг тегов `crud` для полей join-структуры (`ParseJoinTableFlags`, `JoinTableTagFlags`)
- **dialect_vars.go** — пакетные переменные-алиасы (`SQLite`, `PostgresSQL`) для удобного доступа к диалектам без импорта пакета `dialect`
- **cmd/crudsgen** — генератор кода для создания типизированных реализаций Table* и Joiner* без reflection

### Infrastructure Layer (Инфраструктурный слой)

- **dialect** — SQL-диалекты (`PostgreSQLDialect`, `SQLiteDialect`)
- **struct_info** — метаданные таблиц и полей (парсинг тегов, извлечение информации)
- **tx_adapter** — адаптеры для работы с БД (`pgx.Conn`, `pgx.Tx`, `*sql.DB`, `*sql.Tx`)
- **defs** — SQL-константы
- **helpers** — вспомогательные функции (casing)

## Потоки данных

1. Пользовательский код вызывает CRUD-методы `Table[ROW]` или join-методы `Joiner[JT]`
2. `Table[ROW]` / `Joiner[JT]` использует `struct_info` для получения метаданных структуры
3. `dialect` формирует SQL-запросы с учётом особенностей БД
4. Система фильтрации (в корневом пакете) строит дерево условий для WHERE-_clause_
5. `tx_adapter` выполняет запросы через `pgx` или `database/sql`
6. Результаты маппятся обратно в структуры через reflection или сгенерированный код

## Генератор кода (cmd/crudsgen)

Генератор создаёт типизированные реализации таблиц и join-структур без reflection:

- Парсит AST Go файлов для извлечения полей структур
- Обрабатывает embedded структуры и аннонимные поля
- **Режим Table* (-table)**: генерирует hard-coded `TableInfo`, helper методы (`scanRefs`, `insertArgs`, `updateArgs`), CRUD методы с прямым доступом к полям
- **Режим Joiner* (-joiner)**: парсит join-структуры (поля которых — другие структуры с тегами `crud`), генерирует hard-coded `JoinTables` с `TableInfo` для каждой под-таблицы, вызывает `MakeJoinerBase` для построения SQL, генерирует `makeRefs`/`applyRefs` без reflection (с typed temp-переменными для pointer-полей)
- **Поддерживает build-теги** для условной компиляции сгенерированного кода

**Результат:** улучшена производительность, устранена зависимость от reflection в сгенерированном коде.

**Важно:** Сгенерированный код:
- Table* реализует интерфейс `cruds.TypedTable[ROW]`
- Joiner* реализует интерфейс `cruds.TypedJoiner[JT]`
- Не использует reflection для доступа к полям (hard-coded подход)
- Joiner* использует `cruds.MakeJoinerBase` для построения SQL на этапе конструирования
- Joiner* создаёт typed temporary переменные для pointer-полей, обходя type assertion после Scan

## Принципы

- Зависимости направлены внутрь (от инфраструктуры к домену)
- Интерфейсы определяются в том пакете, где используются
- Breaking changes в интерфейсах требуют согласования
