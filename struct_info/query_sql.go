package struct_info

import (
	"strings"

	"github.com/mirrorru/crudquick/defs"
	"github.com/mirrorru/crudquick/dialect"
)

// QuerySqlTexts SQL тексты для Query[T].
// EN: QuerySqlTexts SQL texts for Query[T].
type QuerySqlTexts struct {
	GetOne    string // SELECT ... FROM ... JOINs ... WHERE pk = $1
	ListStart string // SELECT ... FROM ... JOINs ...
	SortPart  string // ORDER BY ...
}

// BuildQuerySqlTexts строит SQL тексты для Query на основе QueryInfo.
// EN: BuildQuerySqlTexts builds SQL texts for Query based on QueryInfo.
func BuildQuerySqlTexts(d dialect.SQLDialect, qi *QueryInfo) QuerySqlTexts {
	return QuerySqlTexts{
		GetOne:    buildQueryGetOneSQL(d, qi),
		ListStart: buildQueryListSQL(qi),
		SortPart:  buildQueryOrderByClause(qi),
	}
}

// buildQueryGetOneSQL строит SQL для One() с WHERE по PK.
// EN: buildQueryGetOneSQL builds SQL for One() with WHERE by PK.
func buildQueryGetOneSQL(d dialect.SQLDialect, qi *QueryInfo) string {
	if len(qi.SelectIdxList) == 0 {
		return ""
	}

	pkTable := qi.Tables[qi.PKIdx]
	if len(pkTable.TableInfo.PKIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	growSize := 100 + len(qi.SelectIdxList)*30 + len(qi.Tables)*50
	sb.Grow(growSize)

	// SELECT
	sb.WriteString(defs.SQLSelect)
	sb.WriteString(buildQueryColumnList(qi))

	// FROM
	sb.WriteString(defs.SQLFrom)
	writeQueryFromClause(&sb, qi)

	// JOINs
	writeQueryJoinClauses(&sb, qi)

	// WHERE по PK
	sb.WriteString(defs.SQLWhere)
	pkOffset := 0
	for pos, pkFieldIdx := range pkTable.TableInfo.PKIdxList {
		if pos > 0 {
			sb.WriteString(defs.SQLAnd)
		}
		// Находим PK поле в combined fields
		pkColName := pkTable.TableInfo.Fields[pkFieldIdx].SQLName
		qualifiedPK := pkTable.Alias + "." + pkColName
		sb.WriteString(qualifiedPK)
		sb.WriteString(defs.SQLEquals)
		sb.WriteString(d.Placeholder(pkOffset + pos + 1))
	}

	return sb.String()
}

// buildQueryListSQL строит SQL для Many() без WHERE (WHERE добавляется фильтром).
// EN: buildQueryListSQL builds SQL for Many() without WHERE (WHERE added by filter).
func buildQueryListSQL(qi *QueryInfo) string {
	if len(qi.SelectIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(100 + len(qi.SelectIdxList)*30 + len(qi.Tables)*50)

	// SELECT
	sb.WriteString(defs.SQLSelect)
	sb.WriteString(buildQueryColumnList(qi))

	// FROM
	sb.WriteString(defs.SQLFrom)
	writeQueryFromClause(&sb, qi)

	// JOINs
	writeQueryJoinClauses(&sb, qi)

	return sb.String()
}

// buildQueryColumnList строит список колонок для SELECT.
// EN: buildQueryColumnList builds column list for SELECT.
func buildQueryColumnList(qi *QueryInfo) string {
	if len(qi.SelectIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(len(qi.SelectIdxList) * 25)
	for i, idx := range qi.SelectIdxList {
		if i > 0 {
			sb.WriteString(defs.SQLCommaSpace)
		}
		sb.WriteString(qi.CombinedFields[idx].SQLName)
	}
	return sb.String()
}

// writeQueryFromClause записывает FROM часть.
// EN: writeQueryFromClause writes FROM clause.
func writeQueryFromClause(sb *strings.Builder, qi *QueryInfo) {
	fromTable := qi.Tables[qi.FromIdx]
	sb.WriteString(fromTable.TableInfo.SQLName)
	// Добавляем AS только если алиас отличается от имени таблицы
	// Add AS only if alias differs from table name
	if fromTable.Alias != fromTable.TableInfo.SQLName {
		sb.WriteString(defs.SQLAs)
		sb.WriteString(fromTable.Alias)
	}
}

// writeQueryJoinClauses записывает JOIN части для всех non-FROM таблиц.
// EN: writeQueryJoinClauses writes JOIN clauses for all non-FROM tables.
func writeQueryJoinClauses(sb *strings.Builder, qi *QueryInfo) {
	for i, qt := range qi.Tables {
		if i == qi.FromIdx {
			continue
		}

		// Определяем тип JOIN
		// Determine JOIN type
		switch qt.JoinType {
		case JoinLeft:
			sb.WriteString(defs.SQLLeftJoin)
		case JoinRight:
			sb.WriteString(defs.SQLRightJoin)
		case JoinInner:
			sb.WriteString(defs.SQLInnerJoin)
		default:
			sb.WriteString(defs.SQLJoin)
		}

		// Имя таблицы
		// Table name
		sb.WriteString(qt.TableInfo.SQLName)
		if qt.Alias != qt.TableInfo.SQLName {
			sb.WriteString(defs.SQLAs)
			sb.WriteString(qt.Alias)
		}

		// ON условия
		// ON conditions
		if len(qt.JoinConds) > 0 {
			sb.WriteString(defs.SQLOn)
			for j, cond := range qt.JoinConds {
				if j > 0 {
					sb.WriteString(defs.SQLAnd)
				}
				// <target_alias>.<target_col> = <source_alias>.<source_col>
				sb.WriteString(cond.TargetAlias)
				sb.WriteString(".")
				sb.WriteString(cond.TargetColumn)
				sb.WriteString(defs.SQLEquals)
				sb.WriteString(qt.Alias)
				sb.WriteString(".")
				sb.WriteString(cond.SourceColumn)
			}
		}
	}
}

// buildQueryOrderByClause строит ORDER BY часть.
// EN: buildQueryOrderByClause builds ORDER BY clause.
func buildQueryOrderByClause(qi *QueryInfo) string {
	if len(qi.SortIdxList) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(20 + len(qi.SortIdxList)*30)
	sb.WriteString(defs.SQLOrderBy)
	for pos, idx := range qi.SortIdxList {
		if pos > 0 {
			sb.WriteString(defs.SQLCommaSpace)
		}
		sb.WriteString(qi.CombinedFields[idx].SQLName)
		if qi.CombinedFields[idx].SortBackward {
			sb.WriteString(defs.SQLDesc)
		}
	}
	return sb.String()
}
