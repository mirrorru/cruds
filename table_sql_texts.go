package quick_crud

import (
	"quick-crud/defs"
	quick_crud "quick-crud/dialect"
	"quick-crud/struct_info"
	"strings"
)

type sqlTexts struct {
	Insert    string
	Update    string
	Delete    string
	GetOne    string
	ListStart string
	SortPart  string
}

var sqlBuilderVal sqlBuilder

type sqlBuilder struct {
}

func (b sqlBuilder) SQLTexts(d quick_crud.SQLDialect, ti *struct_info.TableInfo) sqlTexts {
	return sqlTexts{
		Insert:    b.buildInsertSQL(d, ti),
		Update:    b.buildUpdateSQL(d, ti),
		Delete:    b.buildDeleteSQL(d, ti),
		GetOne:    b.buildGetOneSQL(d, ti),
		ListStart: b.buildListSQL(d, ti),
		SortPart:  buildOrderByClause(ti),
	}
}

func (b sqlBuilder) buildGetOneSQL(d quick_crud.SQLDialect, ti *struct_info.TableInfo) string {
	if len(ti.SelectIdxList) == 0 || len(ti.PKIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	growSize := 50 + len(ti.SelectIdxList)*25 + len(ti.PKIdxList)*25
	sb.Grow(growSize)
	sb.WriteString(defs.SQLSelect)
	sb.WriteString(buildColumnList(ti, ti.SelectIdxList))
	sb.WriteString(defs.SQLFrom)
	sb.WriteString(ti.SQLName)
	b.writeWhereClauses(0, &sb, d, ti)
	return sb.String()
}

func (sqlBuilder) buildListSQL(d quick_crud.SQLDialect, ti *struct_info.TableInfo) string {
	if len(ti.SelectIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(50 + len(ti.SelectIdxList)*25 + len(ti.SQLName))
	sb.WriteString(defs.SQLSelect)
	sb.WriteString(buildColumnList(ti, ti.SelectIdxList))
	sb.WriteString(defs.SQLFrom)
	sb.WriteString(ti.SQLName)
	return sb.String()
}

func (sqlBuilder) buildInsertSQL(d quick_crud.SQLDialect, ti *struct_info.TableInfo) string {
	if len(ti.InsertIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(50 + len(ti.Fields)*50)
	sb.WriteString(defs.SQLInsertInto)
	sb.WriteString(ti.SQLName)
	sb.WriteString(defs.SQLOpenParen)
	sb.WriteString(buildColumnList(ti, ti.InsertIdxList))
	sb.WriteString(defs.SQLCloseParen)
	sb.WriteString(defs.SQLValues)
	sb.WriteString(defs.SQLOpenParen)
	for pos := range ti.InsertIdxList {
		if pos > 0 {
			sb.WriteString(defs.SQLCommaSpace)
		}
		sb.WriteString(d.Placeholder(pos + 1))
	}
	sb.WriteString(defs.SQLCloseParen)

	if d.SupportsReturning() {
		sb.WriteString(defs.SQLReturning)
		sb.WriteString(buildColumnList(ti, ti.SelectIdxList))
	}

	return sb.String()
}

func (b sqlBuilder) buildUpdateSQL(d quick_crud.SQLDialect, ti *struct_info.TableInfo) string {
	if len(ti.UpdateIdxList) == 0 || len(ti.PKIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	growSize := 50 + len(ti.UpdateIdxList)*25 + len(ti.PKIdxList)*25
	if d.SupportsReturning() {
		growSize += len(ti.SelectIdxList) * 25
	}
	sb.Grow(growSize)
	sb.WriteString(defs.SQLUpdate)
	sb.WriteString(ti.SQLName)
	sb.WriteString(defs.SQLSet)
	for pos, idx := range ti.UpdateIdxList {
		if pos > 0 {
			sb.WriteString(defs.SQLCommaSpace)
		}
		sb.WriteString(ti.Fields[idx].SQLName)
		sb.WriteString(defs.SQLEquals)
		sb.WriteString(d.Placeholder(pos + 1))
	}
	b.writeWhereClauses(len(ti.UpdateIdxList), &sb, d, ti)
	if d.SupportsReturning() {
		sb.WriteString(defs.SQLReturning)
		sb.WriteString(buildColumnList(ti, ti.SelectIdxList))
	}

	return sb.String()
}

func (b sqlBuilder) buildDeleteSQL(d quick_crud.SQLDialect, ti *struct_info.TableInfo) string {
	if len(ti.PKIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(30 + len(ti.SQLName) + len(ti.PKIdxList)*25)
	sb.WriteString(defs.SQLDelete)
	sb.WriteString(ti.SQLName)
	b.writeWhereClauses(0, &sb, d, ti)
	return sb.String()
}

func (sqlBuilder) writeWhereClauses(offset int, sb *strings.Builder, d quick_crud.SQLDialect, ti *struct_info.TableInfo) {
	sb.WriteString(defs.SQLWhere)
	for pos, idx := range ti.PKIdxList {
		if pos > 0 {
			sb.WriteString(defs.SQLAnd)
		}
		sb.WriteString(ti.Fields[idx].SQLName)
		sb.WriteString(defs.SQLEquals)
		sb.WriteString(d.Placeholder(offset + pos + 1))
	}
}

func buildColumnList(ti *struct_info.TableInfo, indexes []int) string {
	if len(indexes) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(indexes) * 20)
	for i, idx := range indexes {
		if i > 0 {
			sb.WriteString(defs.SQLCommaSpace)
		}
		sb.WriteString(ti.Fields[idx].SQLName)
	}
	return sb.String()
}

func buildOrderByClause(ti *struct_info.TableInfo) string {
	if len(ti.SortIdxList) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(20 + len(ti.SortIdxList)*25)
	sb.WriteString(defs.SQLOrderBy)
	for pos, idx := range ti.SortIdxList {
		if pos > 0 {
			sb.WriteString(defs.SQLCommaSpace)
		}
		sb.WriteString(ti.Fields[idx].SQLName)
		if ti.Fields[idx].SortBackward {
			sb.WriteString(defs.SQLDesc)
		}
	}
	return sb.String()
}
