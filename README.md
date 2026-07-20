# CRUDs

Библиотека для упрощения CRUD-операций со структурами на Go с использованием дженериков.

[English version](README.EN.md)

## Возможности

- **CRUD-операции**: вставка, обновление, выборка по PK, удаление, выборка с фильтром
- **Поддержка БД**: PostgreSQL, SQLite (обе с поддержкой RETURNING)
- **Система тегов**: настройка поведения полей через struct tag `crud`
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
	"database/sql"
	"fmt"

	qc "github.com/mirrorru/cruds"
	"github.com/mirrorru/cruds/dbtx"
	_ "modernc.org/sqlite"
)

type UserRow struct {
	ID   int64  `crud:"pk;auto"`
	Name string `crud:"col=name"`
}

//func (UserRow) SQLName() string {
//	return "user_table"
//}

func main() {
	// Создание таблицы с диалектом SQLite
	userTable := qc.NewTable[UserRow](qc.SQLite)

	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()

	tx := dbtx.NewDBAdapterVal(db)
	// tx := dbtx.NewTxAdapterVal(dbTrans)
	// tx := dbtx.NewPGXAdapterVal(...)
	// tx := dbtx.NewPGXPoolAdapterVal(...)

	_, _ = tx.ExecContext(context.Background(),
		"CREATE TABLE user_row (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL)")

	// Вставка
	user := &UserRow{Name: "Alice"}
	inserted, res, err := userTable.Ins(context.Background(), tx, user)
	fmt.Printf("Ins: %#v | %#v | %v\n", *inserted, res, err)

	// Выборка по первичному ключу
	found, err := userTable.One(context.Background(), tx, inserted.ID)
	fmt.Printf("One: %#v %v\n", *found, err)

	// Обновление
	inserted.Name = "Bob"
	updated, _, err := userTable.Upd(context.Background(), tx, inserted)
	fmt.Printf("Upd: %#v %v\n", *updated, err)

	// Выборка по фильтру, если требуется
	many, err := userTable.Many(context.Background(), tx, (*qc.Filter)(nil))
	fmt.Printf("Many: %#v %v\n", *many[0], err)

	// Удаление
	_, err = userTable.Del(context.Background(), tx, updated.ID)
	fmt.Println("Del:", err)
}

```

### Алиасы диалектов

В корневом пакете доступны пакетные переменные для удобного доступа к диалектам:

- `qc.PostgresSQL` — PostgreSQL диалект
- `qc.SQLite` — SQLite диалект

## Теги struct полей (`crud:"..."`)

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
| `dbtx` | Адаптеры для `pgx` и `database/sql` |
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
