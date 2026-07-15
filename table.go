package quick_crud

import (
	"context"
	"errors"
	"quick-crud/defs"
	"quick-crud/dialect"
	"quick-crud/filter"
	"quick-crud/struct_info"
	"reflect"
	"strings"

	"github.com/mirrorru/dot"
)

type Table[ROW any] struct {
	dialect   dialect.SQLDialect
	tableInfo struct_info.TableInfo
	sqlTexts  struct_info.SqlTexts
}

func NewTableVal[ROW any](d dialect.SQLDialect) Table[ROW] {
	tableInfo := dot.MustMake(struct_info.GetTableInfo(reflect.TypeFor[ROW]()))

	return Table[ROW]{
		dialect:   d,
		tableInfo: tableInfo,
		sqlTexts:  struct_info.SqlBuilderVal.SQLTexts(d, &tableInfo),
	}
}

func NewTable[ROW any](d dialect.SQLDialect) *Table[ROW] {
	return new(NewTableVal[ROW](d))
}

type tableInternals struct {
	TableInfo struct_info.TableInfo
	SqlTexts  struct_info.SqlTexts
}

func (t *Table[ROW]) Internals() tableInternals {
	return tableInternals{
		TableInfo: t.tableInfo,
		SqlTexts:  t.sqlTexts,
	}
}

func (t *Table[ROW]) Ins(ctx context.Context, tx TxProcessor, row *ROW) (*ROW, Result, error) {
	args := t.tableInfo.Fields.ExtractArgs(row, t.tableInfo.InsertIdxList)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Insert, args)
		return row, sqlResult, err
	}
	buf := new(ROW)
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Insert, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table[ROW]) Upd(ctx context.Context, tx TxProcessor, row *ROW) (*ROW, Result, error) {
	args := t.tableInfo.Fields.ExtractArgs(row, t.tableInfo.UpdateIdxList)
	args = append(args,
		t.tableInfo.Fields.ExtractArgs(row, t.tableInfo.PKIdxList)...,
	)
	if !t.dialect.SupportsReturning() {
		sqlResult, err := tx.ExecContext(ctx, t.sqlTexts.Update, args)
		return row, sqlResult, err
	}
	buf := new(ROW)
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)
	err := tx.QueryRowContext(ctx, t.sqlTexts.Update, args...).Scan(refs...)

	return buf, nil, err
}

func (t *Table[ROW]) One(ctx context.Context, tx TxProcessor, keys ...any) (*ROW, error) {
	buf := new(ROW)
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)
	err := tx.QueryRowContext(ctx, t.sqlTexts.GetOne, keys...).Scan(refs...)

	return buf, err
}

func (t *Table[ROW]) Del(ctx context.Context, tx TxProcessor, keys ...any) (Result, error) {
	return tx.ExecContext(ctx, t.sqlTexts.Delete, keys...)
}

func (t *Table[ROW]) Many(ctx context.Context, tx TxProcessor, filter *filter.Filter) (result []*ROW, err error) {
	var (
		query strings.Builder
		args  []any
		where string
	)
	query.WriteString(t.sqlTexts.ListStart)

	if filter != nil {
		if filter.Range != nil {
			var argIdx int
			where, args, err = filter.Range.Build(t.tableInfo.Fields, t.dialect, &argIdx)
			if err != nil {
				return nil, err
			}
			query.WriteString(defs.SQLWhere)
			query.WriteString(where)
		}
	}
	query.WriteString(t.sqlTexts.SortPart)
	if filter != nil {
		query.WriteString(t.dialect.OffsetAndLimit(filter.Offset, filter.Limit))
	}
	buf := new(ROW)
	refs := t.tableInfo.Fields.ExtractRefs(buf, t.tableInfo.SelectIdxList)

	rows, err := tx.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	for rows.Next() {
		if err = rows.Scan(refs...); err != nil {
			return nil, err
		}
		rec := new(ROW)
		*rec = *buf
		result = append(result, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
