package crudquick

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/mirrorru/crudquick/defs"
	"github.com/mirrorru/crudquick/dialect"
	"github.com/mirrorru/crudquick/struct_info"
)

type Joiner[JT any] struct {
	// JT - joined tables
	tsType     reflect.Type
	joinTables []*JoinTable
	joinerSQLs
}

type joinerSQLs struct {
	GetManySQL string
	GetOneSQL  string
	SortSQL    string
}

func (j *Joiner[JT]) Tables() []*JoinTable {
	return j.joinTables
}

func (j *Joiner[JT]) SQLs() joinerSQLs {
	return j.joinerSQLs
}

func (j *Joiner[JT]) makeRefs() (*JT, []any) {
	result := new(JT)
	return result, nil
}

func (j *Joiner[JT]) One(ctx context.Context, tx TxProcessor, keys ...any) (*JT, error) {
	result, refs := j.makeRefs()
	err := tx.QueryRowContext(ctx, j.GetOneSQL, keys...).Scan(refs...)
	
	return result, err
}

func NewJoiner[JT any](d dialect.SQLDialect) (*Joiner[JT], error) {
	val, err := NewJoinerVal[JT](d)
	if err != nil {
		return nil, err
	}
	return new(val), nil
}

type AllFieldsItem struct {
	tableIdx int
	fieldIdx int
}

func NewJoinerVal[JT any](d dialect.SQLDialect) (Joiner[JT], error) {
	joinTables, err := collectJoinTables(reflect.TypeFor[JT]())
	if err != nil {
		return Joiner[JT]{}, err
	}
	var aliasCnt int
	aliasPosMap := make(map[string]int)
	aliasNameMap := make(map[string]string) // table alias -> real table name
	realNameAliases := make(map[string]string)
	pkTables := make([]int, 1)
	fromIdx := 0
	totalFltCnt := 0
	sortPriotityIdx := make([]int, 0, 1)
	for idx, tInfo := range joinTables {
		if tInfo.isFrom {
			if fromIdx != 0 {
				return Joiner[JT]{}, fmt.Errorf("joiner can't use `from` more then once for tables %d and %d",
					fromIdx, idx)
			}
			fromIdx = idx
		} else if tInfo.isPK && !slices.Contains(pkTables, idx) {
			pkTables = append(pkTables, idx)
		}

		if tInfo.alias == "" {
			for {
				aliasCnt++
				tInfo.alias = fmt.Sprintf("T%d", aliasCnt)
				if _, ok := aliasPosMap[tInfo.alias]; !ok {
					break
				}
			}
		}
		if aliasIdx, ok := aliasPosMap[tInfo.alias]; ok {
			return Joiner[JT]{}, fmt.Errorf("joiner can't use `alias` more then once for tables %d and %d", aliasIdx, idx)
		}
		aliasNameMap[tInfo.alias] = tInfo.tableInfo.SQLName
		realNameAliases[tInfo.tableInfo.SQLName] = tInfo.alias
		totalFltCnt += len(tInfo.tableInfo.SelectIdxList)

		if tInfo.sortPriority != 0 {
			sortPriotityIdx = append(sortPriotityIdx, idx)
		}
	}
	if len(sortPriotityIdx) == 0 {
		sortPriotityIdx = append(sortPriotityIdx, fromIdx)
	} else {
		slices.SortStableFunc(sortPriotityIdx, func(a, b int) int {
			return joinTables[a].sortPriority - joinTables[b].sortPriority
		})
	}
	// Query build
	var selSb strings.Builder
	selSb.Grow(totalFltCnt * 25)
	pos := 0
	selSb.WriteString(defs.SQLSelect)
	for _, tInfo := range joinTables {
		for _, fIdx := range tInfo.tableInfo.SelectIdxList {
			if pos > 0 {
				selSb.WriteString(defs.SQLCommaSpace)
			}
			pos++
			selSb.WriteString(tInfo.alias)
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(tInfo.tableInfo.Fields[fIdx].SQLName)
		}
	}

	selSb.WriteString(defs.SQLFrom)
	selSb.WriteString(joinTables[fromIdx].tableInfo.SQLName)
	selSb.WriteString(defs.SQLAs)
	selSb.WriteString(joinTables[fromIdx].alias)
	defaultJoin := InnerJoin
	if joinTables[fromIdx].isPointer {
		defaultJoin = OuterJoin
	}
	for idx, tInfo := range joinTables {
		if idx == fromIdx {
			continue
		}

		joinMode := tInfo.joinMode
		if joinMode == DefaultJoin {
			joinMode = defaultJoin
		}
		selSb.WriteString("\n")
		selSb.WriteString(joinMode.SQLName())
		selSb.WriteString(defs.SQLSpace)
		selSb.WriteString(tInfo.tableInfo.SQLName)
		selSb.WriteString(defs.SQLAs)
		selSb.WriteString(tInfo.alias)
		selSb.WriteString(defs.SQLOn)
		if len(tInfo.tableInfo.RefIdxList) == 0 {
			selSb.WriteString(defs.SQLTrue)
		}
		for rPos, rIdx := range tInfo.tableInfo.RefIdxList {
			if rPos > 0 {
				selSb.WriteString(defs.SQLAnd)
			}
			selSb.WriteString(joinTables[idx].alias)
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(tInfo.tableInfo.Fields[rIdx].SQLName)
			selSb.WriteString(defs.SQLEquals)
			refTable := tInfo.tableInfo.Fields[rIdx].RefTable
			if refAlias, exists := tInfo.refAliasMap[refTable]; exists {
				refTable = refAlias
			} else {
				refTable = realNameAliases[refTable]
			}
			selSb.WriteString(refTable)
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(tInfo.tableInfo.Fields[rIdx].RefField)
		}
	}
	selSb.WriteString(defs.SQLWhere)
	getManySQL := selSb.String()

	pkTables[0] = fromIdx
	phNum := 0
	for _, pkIdx := range pkTables {
		pkTable := joinTables[pkIdx]
		for _, fldIdx := range pkTable.tableInfo.PKIdxList {
			if phNum > 0 {
				selSb.WriteString(defs.SQLAnd)
			}
			selSb.WriteString(joinTables[pkIdx].alias)
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(pkTable.tableInfo.Fields[fldIdx].SQLName)
			selSb.WriteString(defs.SQLEquals)
			selSb.WriteString(d.Placeholder(phNum))
			phNum++
		}
	}

	// sort
	var sbSort strings.Builder
	sbSort.WriteString(defs.SQLOrderBy)
	sortCnt := 0
	for _, tblIdx := range sortPriotityIdx {
		sortTbl := joinTables[tblIdx]
		for _, fldIdx := range sortTbl.tableInfo.SortIdxList {
			if sortCnt > 0 {
				sbSort.WriteString(defs.SQLCommaSpace)
			}
			fld := sortTbl.tableInfo.Fields[fldIdx]
			sbSort.WriteString(sortTbl.alias)
			sbSort.WriteString(defs.SQLDot)
			sbSort.WriteString(fld.SQLName)
			if fld.SortBackward {
				sbSort.WriteString(defs.SQLDesc)
			}
			sortCnt++
		}
	}
	var sortSQL string
	if sortCnt > 0 {
		sortSQL = sbSort.String()
	}
	return Joiner[JT]{
		tsType:     nil,
		joinTables: joinTables,
		joinerSQLs: joinerSQLs{
			GetManySQL: getManySQL,
			GetOneSQL:  selSb.String(),
			SortSQL:    sortSQL,
		},
	}, nil
}

type JoinMode int32

const (
	DefaultJoin JoinMode = iota
	InnerJoin
	OuterJoin
	LeftJoin
	RightJoin
	CrossJoin
)

func JoinModeParse(in string) (m JoinMode, err error) {
	switch strings.ToLower(in) {
	case "":
		m = DefaultJoin
	case "inner":
		m = InnerJoin
	case "outer":
		m = OuterJoin
	case "left":
		m = LeftJoin
	case "right":
		m = RightJoin
	case "cross":
		m = CrossJoin
	default:
		err = fmt.Errorf("invalid join mode: %s", in)
	}
	return m, err
}

func (j JoinMode) SQLName() string {
	switch j {
	case OuterJoin:
		return "OUTER JOIN"
	case LeftJoin:
		return "LEFT JOIN"
	case RightJoin:
		return "RIGHT JOIN"
	case CrossJoin:
		return "CROSS JOIN"
	default:
		return "JOIN"
	}
}

type JoinTable struct {
	tableInfo    *struct_info.TableInfo
	tableType    reflect.Type
	index        []int
	isPointer    bool
	isPK         bool
	isFrom       bool
	sortPriority int
	joinMode     JoinMode
	refAliasMap  map[string]string
	alias        string
}

func collectJoinTables(t reflect.Type) (result []*JoinTable, err error) {
	if t.Kind() != reflect.Struct {
		return nil, errors.New("expects a struct's type argument for joining")
	}
	result = make([]*JoinTable, 0, t.NumField())
	for idx := range t.NumField() {
		if !t.Field(idx).IsExported() {
			continue
		}
		if t.Field(idx).Anonymous {
			var subRes []*JoinTable
			if subRes, err = collectJoinTables(t.Field(idx).Type); err != nil {
				return nil, err
			}
			for subPos := range subRes {
				subRes[subPos].index = append(t.Field(idx).Index, subRes[subPos].index...)
			}
			result = append(result, subRes...)

			continue
		}

		joinTableFlags, processable := parseJoinTableFlags(t.Field(idx).Tag.Get(struct_info.TagName))
		if !processable {
			continue
		}

		tableType, isPtr := t.Field(idx).Type, false
		if tableType.Kind() == reflect.Ptr {
			tableType, isPtr = tableType.Elem(), true
		}
		if tableType.Kind() != reflect.Struct {
			return nil, errors.New("expects a struct-based type argument")
		}

		var joinMode JoinMode
		if joinTableFlags.Join != "" {
			if joinMode, err = JoinModeParse(joinTableFlags.Join); err != nil {
				return nil, err
			}
		}

		var sortPriority int
		if joinTableFlags.Sort != "" {
			if sortPriority, err = strconv.Atoi(joinTableFlags.Sort); err != nil {
				return nil, err
			}

		}

		var aliasMap map[string]string
		if joinTableFlags.Map != "" {
			maps := strings.Split(joinTableFlags.Map, struct_info.InKeySeparator)
			aliasMap = make(map[string]string, len(maps))
			for _, m := range maps {
				kv := strings.Split(m, struct_info.InKVSeparator)
				if len(kv) != 2 {
					return nil, fmt.Errorf("expecting mapping key-value pairs in '%s'", joinTableFlags.Map)
				}
				aliasMap[kv[0]] = kv[1]
			}
		}

		var tableInfo *struct_info.TableInfo
		if tableInfo, err = struct_info.GetTableInfo(tableType); err != nil {
			return nil, err
		}

		joinTable := &JoinTable{
			tableInfo:    tableInfo,
			tableType:    tableType,
			isPointer:    isPtr,
			isPK:         joinTableFlags.IsPK,
			isFrom:       joinTableFlags.IsFrom,
			sortPriority: sortPriority,
			joinMode:     joinMode,
			refAliasMap:  aliasMap,
			index:        []int{idx},
			alias:        joinTableFlags.Alias,
		}

		result = append(result, joinTable)
	}

	return result, nil
}
