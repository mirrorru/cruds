# CRUDs

A library to simplify CRUD operations with Go structs using generics.

[Русская версия](README.md)

## Features

- **CRUD operations**: insert, update, select by PK, delete, select with filter
- **Database support**: PostgreSQL, SQLite (both with RETURNING support)
- **Tag system**: field behavior configuration via `tbl` struct tag
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
    qc "github.com/mirrorru/cruds"
)

type UserRow struct {
    ID   int64  `tbl:"pk,auto"`
    Name string `tbl:"col=name"`
}

func main() {
    // Create a table with PostgreSQL dialect
    userTable := qc.NewTable[UserRow](qc.PostgresSQL)

    // Insert
    user := &UserRow{Name: "Alice"}
    inserted, _, err := userTable.Ins(context.Background(), tx, user)

    // Select by primary key
    found, err := userTable.One(context.Background(), tx, 1)

    // Update
    inserted.Name = "Bob"
    updated, _, err := userTable.Upd(context.Background(), tx, inserted)

    // Delete
    _, err = userTable.Del(context.Background(), tx, 1)
}
```

### Dialect Aliases

The root package provides package-level variables for convenient dialect access:

- `qc.PostgresSQL` — PostgreSQL dialect
- `qc.SQLite` — SQLite dialect

## Struct Field Tags (`tbl:"..."`)

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
| `tx_adapter` | Adapters for `pgx` and `database/sql` |
| `defs` | SQL constants |
| `helpers` | Utility functions |
| `cmd/crudsgen` | Typed implementation generator |

## Code Generator (crudsgen)

The generator creates typed table implementations without reflection:

```bash
go install ./cmd/crudsgen
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
