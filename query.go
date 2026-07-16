package crudquick

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/mirrorru/crudquick/defs"
	"github.com/mirrorru/crudquick/dialect"
	"github.com/mirrorru/crudquick/struct_info"

	"github.com/mirrorru/dot"
)

// Query мульти-табличный запрос для структур с полями-ROW.
// EN: Query multi-table query for structs with ROW fields.
type Query[T any] struct {
	dialect   dialect.SQLDialect
	queryInfo struct_info.QueryInfo
	sqlTexts  struct_info.QuerySqlTexts
}

// NewQueryVal создает Query[T] по значению.
// EN: NewQueryVal creates Query[T] by value.
func NewQueryVal[T any](d dialect.SQLDialect) Query[T] {
	qi := dot.MustMake(struct_info.CollectQueryInfo(reflect.TypeFor[T]()))

	return Query[T]{
		dialect:   d,
		queryInfo: qi,
		sqlTexts:  struct_info.BuildQuerySqlTexts(d, &qi),
	}
}

// NewQuery создает *Query[T].
// EN: NewQuery creates *Query[T].
func NewQuery[T any](d dialect.SQLDialect) *Query[T] {
	return new(NewQueryVal[T](d))
}

type queryInternals struct {
	QueryInfo      struct_info.QueryInfo
	SqlTexts       struct_info.QuerySqlTexts
	CombinedFields struct_info.TableFields
}

// Internals возвращает внутреннюю метадату Query.
// EN: Internals returns Query internal metadata.
func (q *Query[T]) Internals() queryInternals {
	return queryInternals{
		QueryInfo:      q.queryInfo,
		SqlTexts:       q.sqlTexts,
		CombinedFields: q.queryInfo.CombinedFields,
	}
}

// One возвращает одну запись по PK.
// EN: One returns one record by PK.
func (q *Query[T]) One(ctx context.Context, tx TxProcessor, keys ...any) (*T, error) {
	row := tx.QueryRowContext(ctx, q.sqlTexts.GetOne, keys...)
	return q.scanRow(row)
}

// Many возвращает список записей с опциональным фильтром.
// EN: Many returns list of records with optional filter.
func (q *Query[T]) Many(ctx context.Context, tx TxProcessor, filter *Filter) ([]*T, error) {
	var (
		query strings.Builder
		args  []any
		where string
		err   error
	)

	query.WriteString(q.sqlTexts.ListStart)

	if filter != nil && filter.Range != nil {
		var argIdx int
		where, args, err = filter.Range.Build(q.queryInfo.CombinedFields, q.dialect, &argIdx)
		if err != nil {
			return nil, err
		}
		if where != "" {
			query.WriteString(defs.SQLWhere)
			query.WriteString(where)
		}
	}

	query.WriteString(q.sqlTexts.SortPart)

	if filter != nil {
		query.WriteString(q.dialect.OffsetAndLimit(filter.Offset, filter.Limit))
	}

	rows, err := tx.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	var result []*T
	for rows.Next() {
		rec, scanErr := q.scanRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// cfValue хранит индекс combined field и отсканированное значение.
// EN: cfValue stores combined field index and scanned value.
type cfValue struct {
	cfIdx int
	val   any
}

// scanRow сканирует строку в T с обработкой NULL для pointer полей.
// EN: scanRow scans row into T with NULL handling for pointer fields.
func (q *Query[T]) scanRow(row Row) (*T, error) {
	// Промежуточный буфер для сканирования
	// Intermediate buffer for scanning
	buf := make([]any, len(q.queryInfo.SelectIdxList))
	refs := make([]any, len(buf))
	for i := range refs {
		refs[i] = &buf[i]
	}

	if err := row.Scan(refs...); err != nil {
		return nil, err
	}

	// Группируем значения по T-полям
	// Group values by T-fields
	tFieldValues := make(map[int][]cfValue)
	for i, val := range buf {
		cfIdx := q.queryInfo.SelectIdxList[i]
		cf := q.queryInfo.CombinedFields[cfIdx]
		tFieldIdx := cf.Index[0]
		tFieldValues[tFieldIdx] = append(tFieldValues[tFieldIdx], cfValue{cfIdx: cfIdx, val: val})
	}

	// Аллоцируем T
	// Allocate T
	result := new(T)
	rv := reflect.ValueOf(result).Elem()

	// Обрабатываем каждое T-поле
	// Process each T-field
	for _, qt := range q.queryInfo.Tables {
		tFieldIdx := qt.FieldIndex
		tField := rv.Field(tFieldIdx)
		if !tField.CanSet() {
			continue
		}

		cfValues := tFieldValues[tFieldIdx]
		if len(cfValues) == 0 {
			continue
		}

		// Проверяем, все ли значения NULL
		// Check if all values are NULL
		allNil := true
		for _, cfv := range cfValues {
			if cfv.val != nil {
				allNil = false
				break
			}
		}

		if qt.IsPointer {
			if allNil {
				tField.Set(reflect.Zero(tField.Type()))
				continue
			}
			if tField.IsNil() {
				tField.Set(reflect.New(tField.Type().Elem()))
			}
			tField = tField.Elem()
		}

		// Устанавливаем значения полей
		// Set field values
		for _, cfv := range cfValues {
			cf := q.queryInfo.CombinedFields[cfv.cfIdx]
			rowFieldIdx := cf.Index[1:]
			rowField := tField.FieldByIndex(rowFieldIdx)

			// Пробуем использовать sql.Scanner, если поле его реализует
			// Try sql.Scanner if field implements it
			if rowField.CanAddr() {
				if scanner, ok := rowField.Addr().Interface().(sql.Scanner); ok {
					if err := scanner.Scan(cfv.val); err != nil {
						return nil, err
					}
					continue
				}
			}

			val := reflect.ValueOf(cfv.val)
			if val.IsValid() && val.Type().ConvertibleTo(rowField.Type()) {
				rowField.Set(val.Convert(rowField.Type()))
			}
		}
	}

	return result, nil
}
