# Описание проекта CRUDs

Библиотека для упрощения реализации CRUD-операций и JOIN-запросов со структурами на Go с использованием дженериков.

**Repo**: github.com/mirrorru/cruds
**Go-модуль**: `cruds` (go.mod инициализирован, Go 1.26.3)
**Язык взаимодействия**: русский

## Описание

CRUDs предоставляет типобезопасный слой для работы с БД через дженерик-структуры `Table[ROW any]` и `Joiner[JT any]`.

### Основные возможности

- **CRUD-операции**: `Ins` (insert), `Upd` (update), `One` (select by PK), `Del` (delete), `Many` (select with filter)
- **JOIN-запросы**: `Joiner[JT]` для SELECT с INNER/LEFT/RIGHT/OUTER/CROSS JOIN нескольких таблиц
- **Поддержка БД**: PostgreSQL (с RETURNING), SQLite (с RETURNING)
- **Удобные алиасы диалектов**: пакетные переменные `SQLite` и `PostgresSQL` в корневом пакете для доступа к диалектам без импорта пакета `dialect`
- **Система тегов**: настройка поведения полей через struct tag `crud`
- **Фильтры**: дерево условий с операторами (AND, OR, NOT, =, <>, >, >=, <, <=, LIKE, ILIKE, IN, IS NULL, IS NOT NULL)
- **Пагинация**: OFFSET/LIMIT
- **Адаптеры**: для `pgx` и `database/sql`
- **Генератор кода**: `crudsgen` создаёт типизированные `Table*` и `Joiner*` без reflection

### Структура пакетов

- `dialect` — SQL-диалекты (`PostgreSQLDialect`, `SQLiteDialect`)
- `struct_info` — метаданные таблиц и полей (парсинг тегов, извлечение информации)
- `dbtx` — адаптеры для `pgx.Conn`, `pgx.Tx`, `*sql.DB`, `*sql.Tx`
- `defs` — SQL-константы
- `helpers` — вспомогательные функции (casing)

### Теги struct полей (`crud:"..."`)

- `pk` — первичный ключ
- `ro` — read-only (только SELECT)
- `auto` — автогенерируемое поле (не вставляется)
- `embed` — встраивание структуры
- `omit` — полное игнорирование поля
- `col=<name>` — имя колонки в БД
- `ins` — принудительная вставка поля
- `upd` — принудительное обновление поля
- `rskip` — пропуск при SELECT
- `prefix=<prefix>` — префикс для вложенных колонок
- `ref=<table>:<field>` — внешний ключ (используется для JOIN условий в Joiner)
- `sort=<pos>[:desc]` — сортировка (позиция и направление)

### Теги полей join-структуры для Joiner (`crud:"..."`)

Для полей-таблиц внутри join-структуры используются дополнительные ключи:
- `from` — таблица-источник (FROM), только одна на структуру
- `pk` — таблица участвует в WHERE-условии для `One()` (по PK)
- `sort=<pos>` — приоритет ORDER BY среди таблиц
- `alias=<name>` — явный алиас таблицы в SQL
- `join=<mode>` — режим JOIN: `inner`, `left`, `right`, `outer`, `cross`
- `map=<table>:<alias>,...` — маппинг псевдонимов для ref-полей

Pointer-поля (`*SomeRow`) в join-структуре автоматически получают OUTER JOIN по умолчанию.
