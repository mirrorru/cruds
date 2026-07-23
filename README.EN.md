# CRUDs

A library to simplify CRUD operations with Go structs using generics.

[Русская версия](README.md)

## Features

- **CRUD operations**: insert, update, select by PK, delete, select with filter
- **Database support**: PostgreSQL, SQLite (both with RETURNING support)
- **Tag system**: field behavior configuration via `crud` struct tag
- **Filters**: condition tree with operators (AND, OR, NOT, =, <>, >, >=, <, <=, LIKE, ILIKE, IN, IS NULL, IS NOT NULL)
- **Pagination**: OFFSET / LIMIT
- **Adapters**: for `pgx` and `database/sql`
- **Code generator**: typed table implementations without reflection

## Installation

```bash
go get github.com/mirrorru/cruds
```

## Quick Start

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
	// Create a table with SQLite dialect
	userTable := qc.NewTable[UserRow](qc.SQLite)

	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()

	tx := dbtx.NewDBAdapterVal(db)
	// tx := dbtx.NewTxAdapterVal(dbTrans)
	// tx := dbtx.NewPGXAdapterVal(...)
	// tx := dbtx.NewPGXPoolAdapterVal(...)

	_, _ = tx.ExecContext(context.Background(),
		"CREATE TABLE user_row (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL)")

	// Insert
	user := &UserRow{Name: "Alice"}
	inserted, res, err := userTable.Ins(context.Background(), tx, user)
	fmt.Printf("Ins: %#v | %#v | %v\n", *inserted, res, err)

	// Select by primary key
	found, err := userTable.One(context.Background(), tx, inserted.ID)
	fmt.Printf("One: %#v %v\n", *found, err)

	// Update
	inserted.Name = "Bob"
	updated, _, err := userTable.Upd(context.Background(), tx, inserted)
	fmt.Printf("Upd: %#v %v\n", *updated, err)

	// Select by filter, or all
	many, err := userTable.Many(context.Background(), tx, (*qc.Filter)(nil))
	fmt.Printf("Many: %#v %v\n", *many[0], err)

	// Delete
	_, err = userTable.Del(context.Background(), tx, updated.ID)
	fmt.Println("Del:", err)
}
```

### Dialect Aliases

The root package provides package-level variables for convenient dialect access:

- `qc.PostgresSQL` — PostgreSQL dialect
- `qc.SQLite` — SQLite dialect

## Struct Field Tags (`crud:"..."`)

| Tag | Description |
|-----|-------------|
| `pk` | Primary key |
| `ro` | Read-only (SELECT only) |
| `auto` | Auto-generated field |
| `embed` | Embedded struct |
| `omit` | Ignore field |
| `col=<name>` | Column name in DB |
| `ins` | Force insert |
| `upd` | Force update |
| `rskip` | Skip on SELECT |
| `prefix=<prefix>` | Prefix for nested columns |
| `ref=<table>,<field>` | Foreign key |
| `sort=<pos>[:desc]` | Sort order |

## Package Structure

| Package | Description |
|---------|-------------|
| `dialect` | SQL dialects (`PostgreSQLDialect`, `SQLiteDialect`) |
| `filter` | Filter system |
| `struct_info` | Table and field metadata |
| `dbtx` | Adapters for `pgx` and `database/sql` |
| `defs` | SQL constants |
| `helpers` | Utility functions |
| `cmd/crudsgen` | Typed implementation generator |

## Code Generator (crudsgen)

The generator creates typed table implementations without reflection:

```bash
go install github.com/mirrorru/cruds/cmd/crudsgen
crudsgen -src=path/to/models:*Row -dest=path/to/repo -pkg=repo
```

## Build & Test

```bash
# Standard check
go vet ./... && go build ./... && go test ./... && task lint

# All tests
task test

# Linter
task lint
```

## License

See [LICENSE](LICENSE) file.
