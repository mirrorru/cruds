# CRUDs

Библиотека для упрощения CRUD-операций со структурами на Go с использованием дженериков.

[English version](README.EN.md)

## Возможности

- **CRUD-операции**: вставка, обновление, выборка по PK, удаление, выборка с фильтром
- **Поддержка БД**: PostgreSQL, SQLite (обе с поддержкой RETURNING)
- **Система тегов**: настройка поведения полей через struct tag `tbl`
- **Фильтры**: дерево условий с операторами (AND, OR, NOT, =, <>, >, >=, <, <=, LIKE, ILIKE, IN, IS NULL, IS NOT NULL)
- **Пагинация**: OFFSET / LIMIT
- **Адаптеры**: для `pgx` и `database/sql`
- **Генератор кода**: типизированные реализации таблиц без reflection

## Установка

```bash
go get github.com/mirrorru/cruds
```

## Быстрый старт

```go
package main

import (
    "context"
    qc "github.com/mirrorru/cruds"
)

type UserRow struct {
    ID   int64  `tbl:"pk,auto"`
    Name string `tbl:"col=name"`
}

func main() {
    // Создание таблицы с диалектом PostgreSQL
    userTable := qc.NewTable[UserRow](qc.PostgresSQL)

    // Вставка
    user := &UserRow{Name: "Alice"}
    inserted, _, err := userTable.Ins(context.Background(), tx, user)

    // Выборка по первичному ключу
    found, err := userTable.One(context.Background(), tx, 1)

    // Обновление
    inserted.Name = "Bob"
    updated, _, err := userTable.Upd(context.Background(), tx, inserted)

    // Удаление
    _, err = userTable.Del(context.Background(), tx, 1)
}
```

### Алиасы диалектов

В корневом пакете доступны пакетные переменные для удобного доступа к диалектам:

- `qc.PostgresSQL` — PostgreSQL диалект
- `qc.SQLite` — SQLite диалект

## Теги struct полей (`tbl:"..."`)

| Тег | Описание |
|-----|----------|
| `pk` | Первичный ключ |
| `ro` | Read-only (только SELECT) |
| `auto` | Автогенерируемое поле |
| `embed` | Встраивание структуры |
| `omit` | Игнорирование поля |
| `col=<name>` | Имя колонки в БД |
| `ins` | Принудительная вставка |
| `upd` | Принудительное обновление |
| `rskip` | Пропуск при SELECT |
| `prefix=<prefix>` | Префикс для вложенных колонок |
| `ref=<table>,<field>` | Внешний ключ |
| `sort=<pos>[:desc]` | Сортировка |

## Структура пакетов

| Пакет | Описание |
|-------|----------|
| `dialect` | SQL-диалекты (`PostgreSQLDialect`, `SQLiteDialect`) |
| `filter` | Система фильтрации |
| `struct_info` | Метаданные таблиц и полей |
| `tx_adapter` | Адаптеры для `pgx` и `database/sql` |
| `defs` | SQL-константы |
| `helpers` | Вспомогательные функции |
| `cmd/crudsgen` | Генератор типизированных реализаций |

## Генератор кода (crudsgen)

Генератор создаёт типизированные реализации таблиц без reflection:

```bash
go install ./cmd/crudsgen
crudsgen -src=path/to/models:*Row -dest=path/to/repo -pkg=repo
```

## Сборка и тесты

```bash
# Стандартная проверка
go vet ./... && go build ./... && go test ./... && task lint

# Все тесты
task test

# Линтер
task lint
```

## Лицензия

См. файл [LICENSE](LICENSE).
