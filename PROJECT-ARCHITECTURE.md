# Архитектура проекта Quick CRUD

## Подход

Domain Driven Design & Clean Architecture.

## Слои архитектуры

### Domain Layer (Доменный слой)

- **contracts** — интерфейсы и контракты (`TxProcessor`, `Row`, `Result`, `Rows`, `TypedTable[ROW]`)
- Определяют абстракции, не зависят от реализаций

### Application Layer (Слой приложения)

- **Корневой пакет (quick-crud)** — универсальная реализация `Table[ROW]` с использованием reflection
- **cmd/qcgen** — генератор кода для создания типизированных реализаций без reflection

### Infrastructure Layer (Инфраструктурный слой)

- **dialect** — SQL-диалекты (`PostgreSQLDialect`, `SQLiteDialect`)
- **filter** — система фильтрации (`Filter`, `FilterNode`, `ConditionNode`, `GroupNode`)
- **struct_info** — метаданные таблиц и полей (парсинг тегов, извлечение информации)
- **tx_adapter** — адаптеры для работы с БД (`pgx.Conn`, `pgx.Tx`, `*sql.DB`, `*sql.Tx`)
- **defs** — SQL-константы
- **helpers** — вспомогательные функции (casing)

## Потоки данных

1. Пользовательский код вызывает CRUD-методы `Table[ROW]`
2. `Table[ROW]` использует `struct_info` для получения метаданных структуры
3. `dialect` формирует SQL-запросы с учётом особенностей БД
4. `filter` строит дерево условий для WHERE-_clause
5. `tx_adapter` выполняет запросы через `pgx` или `database/sql`
6. Результаты маппятся обратно в структуры через reflection или сгенерированный код

## Генератор кода (cmd/qcgen)

Генератор создаёт типизированные реализации таблиц без reflection:

- Парсит AST Go файлов для извлечения полей структур
- Обрабатывает embedded структуры
- Генерирует hard-coded `TableInfo` вместо вызовов `GetTableInfo(reflect.TypeFor[...])`
- Генерирует helper методы: `scanRefs`, `insertArgs`, `updateArgs`
- CRUD методы используют прямой доступ к полям

**Результат:** улучшена производительность, устранена зависимость от reflection в сгенерированном коде.

## Принципы

- Зависимости направлены внутрь (от инфраструктуры к домену)
- Интерфейсы определяются в том пакете, где используются
- Breaking changes в интерфейсах требуют согласования
