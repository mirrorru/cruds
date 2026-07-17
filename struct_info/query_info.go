package struct_info

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
)

// JoinCondition описывает условие ON для JOIN.
// EN: JoinCondition describes an ON condition for JOIN.
type JoinCondition struct {
	TargetAlias  string // алиас referenced таблицы / alias of referenced table
	TargetColumn string // referenced column name
	SourceColumn string // колонка в source таблице с ref= / column in source table with ref=
}

const (
	// JoinLeft LEFT JOIN тип.
	JoinLeft = "left"
	// JoinRight RIGHT JOIN тип.
	JoinRight = "right"
	// JoinInner INNER JOIN тип.
	JoinInner = "inner"
)

// QueryTableInfo метаинформация для одной таблицы в Query.
// EN: QueryTableInfo metadata for one table in Query.
type QueryTableInfo struct {
	TableInfo  *TableInfo
	Alias      string
	JoinType   string // "left", "right", "inner"
	IsFrom     bool
	IsPK       bool
	IsPointer  bool
	FieldIndex int
	SortPos    int
	RefMap     map[string]string
	JoinConds  []JoinCondition
}

// QueryInfo полная метаинформация для Query[T].
// EN: QueryInfo full metadata for Query[T].
type QueryInfo struct {
	Tables         []QueryTableInfo
	FromIdx        int
	PKIdx          int
	CombinedFields TableFields
	SelectIdxList  []int
	SortIdxList    []int
	FieldNameIdx   map[string]int
}

var collectQueryInfoCache sync.Map

// CollectQueryInfo собирает метаинформацию для Query-структуры T.
// EN: CollectQueryInfo collects metadata for Query struct T.
func CollectQueryInfo(t reflect.Type) (QueryInfo, error) {
	if t.Kind() != reflect.Struct {
		return QueryInfo{}, errors.New("CollectQueryInfo: expects a struct type")
	}

	if v, ok := collectQueryInfoCache.Load(t); ok {
		return v.(QueryInfo), nil //nolint:errcheck
	}

	qi, err := collectQueryInfoImpl(t)
	if err != nil {
		return QueryInfo{}, err
	}

	collectQueryInfoCache.Store(t, qi)
	return qi, nil
}

func collectQueryInfoImpl(t reflect.Type) (QueryInfo, error) {
	result := QueryInfo{
		Tables:       make([]QueryTableInfo, 0, t.NumField()),
		FieldNameIdx: make(map[string]int),
	}

	// Шаг 1-6: собираем QueryTableInfo для каждого поля T
	// Step 1-6: collect QueryTableInfo for each T field
	for idx := range t.NumField() {
		fld := t.Field(idx)
		if !fld.IsExported() {
			continue
		}

		flags, err := ParseQueryTag(fld.Tag.Get(TagName))
		if err != nil {
			return QueryInfo{}, fmt.Errorf("field %s: %w", fld.Name, err)
		}
		if flags.IsOmit {
			continue
		}

		// Определяем тип ROW (разыменовываем pointer если нужно)
		// Determine ROW type (dereference pointer if needed)
		fieldType := fld.Type
		isPointer := fieldType.Kind() == reflect.Pointer
		if isPointer {
			fieldType = fieldType.Elem()
		}

		// Получаем TableInfo для ROW типа
		// Get TableInfo for ROW type
		tableInfo, err := GetTableInfo(fieldType)
		if err != nil {
			return QueryInfo{}, fmt.Errorf("field %s: %w", fld.Name, err)
		}

		// Определяем алиас (явный или авто)
		// Determine alias (explicit or auto)
		alias := flags.Alias
		if alias == "" {
			alias = tableInfo.SQLName
		}

		qt := QueryTableInfo{
			TableInfo:  tableInfo,
			Alias:      alias,
			IsFrom:     flags.IsFrom,
			IsPK:       flags.IsPK,
			IsPointer:  isPointer,
			FieldIndex: idx,
			SortPos:    flags.SortPos,
			RefMap:     flags.RefMap,
			JoinType:   flags.JoinType,
		}

		result.Tables = append(result.Tables, qt)
	}

	if len(result.Tables) == 0 {
		return QueryInfo{}, errors.New("CollectQueryInfo: no non-omit fields in Query struct")
	}

	// Шаг 7: определяем FROM таблицу
	// Step 7: determine FROM table
	result.FromIdx = -1
	for i, qt := range result.Tables {
		if qt.IsFrom {
			if result.FromIdx != -1 {
				return QueryInfo{}, errors.New("CollectQueryInfo: multiple FROM tables")
			}
			result.FromIdx = i
		}
	}
	if result.FromIdx == -1 {
		result.FromIdx = 0 // первая non-omit таблица / first non-omit table
		result.Tables[0].IsFrom = true
	}

	// Шаг 8: определяем PK таблицу
	// Step 8: determine PK table
	result.PKIdx = -1
	for i, qt := range result.Tables {
		if qt.IsPK {
			if result.PKIdx != -1 {
				return QueryInfo{}, errors.New("CollectQueryInfo: multiple PK tables")
			}
			result.PKIdx = i
		}
	}
	if result.PKIdx == -1 {
		result.PKIdx = result.FromIdx // по умолчанию = FROM таблица / default = FROM table
		result.Tables[result.PKIdx].IsPK = true
	}

	// Шаг 9: определяем default JOIN типы
	// Step 9: determine default JOIN types
	fromIsPointer := result.Tables[result.FromIdx].IsPointer
	for i := range result.Tables {
		if i == result.FromIdx {
			continue
		}
		qt := &result.Tables[i]
		if qt.JoinType == "" {
			// Auto-determine JOIN type
			switch {
			case fromIsPointer:
				qt.JoinType = JoinLeft
			case qt.IsPointer:
				qt.JoinType = JoinLeft
			default:
				qt.JoinType = JoinInner
			}
		}
	}

	// Шаг 10: собираем JOIN условия
	// Step 10: collect JOIN conditions
	if err := collectJoinConditions(&result); err != nil {
		return QueryInfo{}, err
	}

	// Шаг 11: строим combined TableFields
	// Step 11: build combined TableFields
	if err := buildCombinedFields(&result); err != nil {
		return QueryInfo{}, err
	}

	// Шаг 12: строим SelectIdxList и SortIdxList
	// Step 12: build SelectIdxList and SortIdxList
	buildSelectAndSortIdxLists(&result)

	return result, nil
}

// collectJoinConditions собирает условия ON для JOIN'ов.
// EN: collectJoinConditions collects ON conditions for JOINs.
func collectJoinConditions(qi *QueryInfo) error {
	// Для каждой таблицы (source) собираем условия из её ref= полей
	// For each table (source) collect conditions from its ref= fields
	for srcIdx, srcQt := range qi.Tables {
		if srcIdx == qi.FromIdx {
			continue // FROM таблица не JOIN'ится / FROM table is not JOINed
		}

		// Проходим по всем ref= полям в этой таблице
		// Iterate over all ref= fields in this table
		for _, refFieldIdx := range srcQt.TableInfo.RefIdxList {
			refField := srcQt.TableInfo.Fields[refFieldIdx]
			refTable := refField.RefTable
			refCol := refField.RefField

			// Резолвим target алиас через RefMap или прямое совпадение
			// Resolve target alias via RefMap or direct match
			targetAlias := ""
			if srcQt.RefMap != nil {
				if mapped, ok := srcQt.RefMap[refTable]; ok {
					targetAlias = mapped
				}
			}
			if targetAlias == "" {
				// Ищем совпадение по алиасам или именам таблиц
				// Look for match by aliases or table names
				for _, tgtQt := range qi.Tables {
					if tgtQt.Alias == refTable || tgtQt.TableInfo.SQLName == refTable {
						targetAlias = tgtQt.Alias
						break
					}
				}
			}
			if targetAlias == "" {
				return fmt.Errorf("CollectQueryInfo: cannot resolve target for ref=%s:%s in table %s",
					refTable, refCol, srcQt.Alias)
			}

			// Создаем JoinCondition и добавляем к source таблице
			// Create JoinCondition and add to source table
			cond := JoinCondition{
				TargetAlias:  targetAlias,
				TargetColumn: refCol,
				SourceColumn: refField.SQLName,
			}
			qi.Tables[srcIdx].JoinConds = append(qi.Tables[srcIdx].JoinConds, cond)
		}
	}

	return nil
}

// buildCombinedFields строит плоский список всех selectable полей из всех таблиц.
// EN: buildCombinedFields builds flat list of all selectable fields from all tables.
func buildCombinedFields(qi *QueryInfo) error {
	combined := make(TableFields, 0, 32)

	for tFieldIdx, qt := range qi.Tables {
		for rowFieldIdx, rowField := range qt.TableInfo.Fields {
			if !rowField.CanSelect {
				continue
			}

			// Квалифицированное имя: alias.column
			// Qualified name: alias.column
			qualifiedName := qt.Alias + "." + rowField.SQLName

			// Проверяем уникальность имени
			// Check name uniqueness
			if prevIdx, ok := qi.FieldNameIdx[qualifiedName]; ok {
				return fmt.Errorf("CollectQueryInfo: duplicate field name %s at indices %d and [%d,%d]",
					qualifiedName, prevIdx, tFieldIdx, rowFieldIdx)
			}

			// Создаем combined field
			// Create combined field
			cf := TableField{
				Index:        append([]int{qt.FieldIndex}, rowField.Index...),
				Path:         slices.Insert(rowField.Path, 0, qi.Tables[tFieldIdx].TableInfo.SQLName),
				SQLName:      qualifiedName,
				RefTable:     rowField.RefTable,
				RefField:     rowField.RefField,
				SortPos:      rowField.SortPos,
				IsPK:         rowField.IsPK,
				CanSelect:    rowField.CanSelect,
				CanInsert:    rowField.CanInsert,
				CanUpdate:    rowField.CanUpdate,
				SortBackward: rowField.SortBackward,
			}

			qi.FieldNameIdx[qualifiedName] = len(combined)
			combined = append(combined, cf)
		}
	}

	qi.CombinedFields = combined
	return nil
}

// buildSelectAndSortIdxLists строит списки индексов для SELECT и ORDER BY.
// EN: buildSelectAndSortIdxLists builds index lists for SELECT and ORDER BY.
func buildSelectAndSortIdxLists(qi *QueryInfo) {
	qi.SelectIdxList = make([]int, 0, len(qi.CombinedFields))
	qi.SortIdxList = make([]int, 0, len(qi.CombinedFields)/4)

	for idx, cf := range qi.CombinedFields {
		if cf.CanSelect {
			qi.SelectIdxList = append(qi.SelectIdxList, idx)
		}
		if cf.SortPos != 0 {
			qi.SortIdxList = append(qi.SortIdxList, idx)
		}
	}
}
