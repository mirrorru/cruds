# Структура проекта Quick CRUD

## Обзор

Quick CRUD - библиотека для упрощения CRUD операций со структурами на Go с использованием generics.

## Структура пакетов

### Корневой пакет (quick-crud)
- `table.go` - универсальная реализация `Table[ROW]` с использованием reflection

### cmd/qcgen
Генератор кода для создания типизированных реализаций таблиц без reflection.
- `main.go` - генератор, который парсит AST Go файлов и создает hard-coded реализации

**Основные возможности генератора:**
- Парсинг полей структуры из AST
- Обработка embedded структур
- Извлечение SQL имени из метода `SQLName()`
- Генерация hard-coded `TableInfo` без reflection
- Генерация helper методов: `scanRefs`, `insertArgs`, `updateArgs`
- Генерация CRUD методов с прямым доступом к полям

### contracts
Интерфейсы и контракты:
- `TxProcessor` - интерфейс для работы с транзакциями
- `TypedTable[ROW]` - интерфейс для типизированных таблиц

### dialect
SQL диалекты:
- `postgres.go` - PostgreSQL диалект
- `sqlite.go` - SQLite диалект

### filter
Система фильтрации:
- `filter.go` - дерево условий (ConditionNode, GroupNode, FilterNode)

### helpers
Вспомогательные функции:
- `casing.go` - преобразование CamelCase в snake_case

### struct_info
Метаданные таблиц и полей:
- `field_tags.go` - парсинг тегов `tbl`
- `table_field.go` - `TableField`, `TableFields`, `ExtractArgs`, `ExtractRefs`
- `table_info.go` - `TableInfo`, `GetTableInfo`
- `table_sql_texts.go` - построение SQL текстов

### tx_adapter
Адаптеры для работы с БД:
- `pgx_adapter.go` - адаптеры для pgx.Conn и pgx.Tx
- `sql_adapter.go` - адаптеры для *sql.DB и *sql.Tx

### test
Тестовые файлы:
- `gen/model/models.go` - тестовые модели (UserRow, ProductRow)
- `gen/repo/` - сгенерированные реализации таблиц
- `smoke/` - smoke тесты
- `samples/` - примеры структур

## Недавние изменения

### Генератор кода (cmd/qcgen)

**Задача:** Заменить вызовы reflection на hard-coded подход в генерируемых реализациях.

**Внесенные изменения:**

1. **Парсинг полей из AST:**
   - Добавлен парсинг полей структуры из AST Go файлов
   - Реализована обработка embedded структур
   - Добавлено извлечение SQL имени из метода `SQLName()`

2. **Hard-coded генерация:**
   - Генерация `TableInfo` как struct literal вместо вызова `GetTableInfo(reflect.TypeFor[...])`
   - Генерация helper методов `scanRefs`, `insertArgs`, `updateArgs` с прямым доступом к полям
   - CRUD методы используют helper методы вместо `ExtractArgs`/`ExtractRefs`

3. **Удалены зависимости от reflection:**
   - Удален импорт `reflect` из генерируемого кода
   - Удален импорт `github.com/mirrorru/dot` из генерируемого кода
   - Удалены вызовы `struct_info.GetTableInfo`
   - Удалены вызовы `t.tableInfo.Fields.ExtractArgs`
   - Удалены вызовы `t.tableInfo.Fields.ExtractRefs`

**Результат:**
- Генерируемый код не использует reflection
- Улучшена производительность (прямой доступ к полям)
- Сохранена совместимость с универсальной реализацией `Table[ROW]`
- Все тесты проходят успешно

## Использование генератора

```bash
# Установка генератора
go install ./cmd/qcgen

# Генерация кода
qcgen -src=path/to/models:*Row -dest=path/to/repo -pkg=repo
```

**Пример:**
```bash
qcgen -src=test/gen/model:*Row -dest=test/gen/repo -pkg=repo
```

## Тестирование

```bash
# Unit тесты
task test:unit

# Smoke тесты
task test:smoke

# Тесты сгенерированного кода
go test -v -tags=smoke ./test/gen/repo/...
```

## Линтинг

```bash
task lint
```

**Примечание:** Оставшиеся предупреждения линтера:
- `gocognit` - высокая когнитивная сложность функций генератора (не критично)
- `lll` - длинные строки в шаблоне (ожидаемо для генерации кода)
- `nestif` - сложные вложенные блоки (не критично)
- `staticcheck` - устаревший `parser.ParseDir` (предупреждение о депрекации)

Все критические ошибки линтера исправлены.
