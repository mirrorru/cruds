# Описание проекта Quick CRUD

Библиотека для упрощения реализации CRUD-операций со структурами на Go с использованием дженериков.

**Repo**: github.com/mirrorru/quick-crud
**Go-модуль**: `quick-crud` (go.mod инициализирован, Go 1.26.3)
**Язык взаимодействия**: русский

## Описание

Quick CRUD предоставляет типобезопасный слой для работы с БД через дженерик-структуру `Table[ROW any]`.

### Основные возможности

- **CRUD-операции**: `Ins` (insert), `Upd` (update), `One` (select by PK), `Del` (delete), `Many` (select with filter)
- **Поддержка БД**: PostgreSQL (с RETURNING), SQLite (с RETURNING)
- **Система тегов**: настройка поведения полей через struct tag `tbl`
- **Фильтры**: дерево условий с операторами (AND, OR, NOT, =, <>, >, >=, <, <=, LIKE, ILIKE, IN, IS NULL, IS NOT NULL)
- **Пагинация**: OFFSET/LIMIT
- **Адаптеры**: для `pgx` и `database/sql`

### Структура пакетов

- `contracts` — интерфейсы (`TxProcessor`, `Row`, `Result`, `Rows`)
- `dialect` — SQL-диалекты (`PostgreSQLDialect`, `SQLiteDialect`)
- `filter` — система фильтрации (`Filter`, `FilterNode`, `ConditionNode`, `GroupNode`)
- `struct_info` — метаданные таблиц и полей (парсинг тегов, извлечение информации)
- `tx_adapter` — адаптеры для `pgx.Conn`, `pgx.Tx`, `*sql.DB`, `*sql.Tx`
- `defs` — SQL-константы
- `helpers` — вспомогательные функции (casing)

### Теги struct полей (`tbl:"..."`)

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
- `ref=<table>,<field>` — внешний ключ
- `sort=<pos>[:desc]` — сортировка (позиция и направление)
