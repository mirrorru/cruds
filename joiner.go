//nolint:gocognit, gocyclo, cyclop, govet, funlen
package cruds

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/mirrorru/cruds/defs"
	"github.com/mirrorru/cruds/dialect"
	"github.com/mirrorru/cruds/struct_info"
)

type JoinTables []JoinTable

type JoinerBase struct {
	dialect    dialect.SQLDialect
	joinTables JoinTables
	allFields  struct_info.TableFields
	fldPosMap  map[string]int
	sql        joinerSQLs
}

type Joiner[JT any] struct {
	// JT - joined tables
	JoinerBase
}

type joinerSQLs struct {
	GetManySQL string
	GetOneSQL  string
	SortSQL    string
}

func (j *Joiner[JT]) Tables() JoinTables {
	return j.joinTables
}

func (j *Joiner[JT]) SQLs() joinerSQLs {
	return j.sql
}

func (jb JoinerBase) AllFields() struct_info.TableFields { return jb.allFields }
func (jb JoinerBase) OneSQL() string                     { return jb.sql.GetOneSQL }
func (jb JoinerBase) ManySQL() string                    { return jb.sql.GetManySQL }
func (jb JoinerBase) SortSQL() string                    { return jb.sql.SortSQL }

func (jts JoinTables) MakeRefs(in any) []any {
	result := make([]any, 0)
	elem := reflect.ValueOf(in).Elem()
	for _, table := range jts {
		if table.IsPointer {
			refs := make([]any, len(table.TableInfo.SelectIdxList))
			for i := range refs {
				result = append(result, &refs[i])
			}
			continue
		}
		tableRef := elem.FieldByIndex(table.Index).Addr().Interface()
		refs := table.TableInfo.Fields.ExtractRefs(tableRef, table.TableInfo.SelectIdxList)
		result = append(result, refs...)
	}
	return result
}

func (jts JoinTables) ApplyRefs(in any, refs []any) {
	pos := 0
	elem := reflect.ValueOf(in).Elem()
	for _, table := range jts {
		if !table.IsPointer {
			pos += len(table.TableInfo.SelectIdxList)
			continue
		}

		filled := false
		checkFrom := pos
		tField := elem.FieldByIndex(table.Index)
		for range table.TableInfo.SelectIdxList {
			p, _ := refs[checkFrom].(*any)
			if *p != nil {
				filled = true
				break
			}
			checkFrom++
		}
		if !filled {
			pos += len(table.TableInfo.SelectIdxList)
			// Сброс pointer-поля в nil перед проверкой, чтобы избежать
			// утечки данных когда связанная таблица возвращает NULL
			tField.SetZero()
			continue
		}
		tField.Set(reflect.New(tField.Type().Elem()))
		tField = tField.Elem()
		for _, fIdx := range table.TableInfo.SelectIdxList {
			p, _ := refs[pos].(*any)
			pos++
			val := *p
			if val == nil {
				continue
			}
			fieldVal := tField.FieldByIndex(table.TableInfo.Fields[fIdx].Index)
			rv := reflect.ValueOf(val)
			if rv.Type().AssignableTo(fieldVal.Type()) {
				fieldVal.Set(rv)
			} else if scanner, ok := fieldVal.Addr().Interface().(sql.Scanner); ok {
				_ = scanner.Scan(val)
			} else if rv.Type().ConvertibleTo(fieldVal.Type()) {
				fieldVal.Set(rv.Convert(fieldVal.Type()))
			}
		}
	}
}

func (j *Joiner[JT]) One(ctx context.Context, tx TxProcessor, keys ...any) (*JT, error) {
	result := new(JT)
	refs := j.joinTables.MakeRefs(result)
	err := tx.QueryRowContext(ctx, j.sql.GetOneSQL, keys...).Scan(refs...)
	j.joinTables.ApplyRefs(result, refs)
	return result, err
}

func MakeQuery4Many(j JoinerBase, filter *Filter) (query string, args []any, err error) {
	var sb strings.Builder
	sb.WriteString(j.sql.GetManySQL)
	if filter != nil {
		if filter.Range != nil {
			var (
				argCnt int
				where  string
			)
			if where, args, err = filter.Range.Build(j.allFields, j.dialect, &argCnt); err != nil {
				return "", nil, err
			}
			sb.WriteString(defs.SQLWhere)
			sb.WriteString(where)
		}
	}
	sb.WriteString(j.sql.SortSQL)
	if filter != nil {
		sb.WriteString(j.dialect.OffsetAndLimit(filter.Offset, filter.Limit))
	}

	return sb.String(), args, nil
}

func (j *Joiner[JT]) Many(ctx context.Context, tx TxProcessor, filter *Filter) (result []*JT, err error) {
	query, args, err := MakeQuery4Many(j.JoinerBase, filter)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	buf := new(JT)
	refs := j.joinTables.MakeRefs(buf)
	for rows.Next() {
		if err = rows.Scan(refs...); err != nil {
			return nil, err
		}
		j.joinTables.ApplyRefs(buf, refs)
		rec := new(JT)
		*rec = *buf
		result = append(result, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, err
}

func NewJoiner[JT any](d dialect.SQLDialect) (*Joiner[JT], error) {
	val, err := NewJoinerVal[JT](d)
	if err != nil {
		return nil, err
	}
	return new(val), nil
}

func NewJoinerVal[JT any](d dialect.SQLDialect) (Joiner[JT], error) {
	joinTables, err := collectJoinTables(reflect.TypeFor[JT]())
	if err != nil {
		return Joiner[JT]{}, err
	}
	jBase, err := MakeJoinerBase(joinTables, d)
	if err != nil {
		return Joiner[JT]{}, err
	}
	return Joiner[JT]{
		JoinerBase: jBase,
	}, nil
}

func MakeJoinerBase(joinTables JoinTables, d dialect.SQLDialect) (JoinerBase, error) {
	var aliasCnt int
	aliasPosMap := make(map[string]int)
	aliasNameMap := make(map[string]string) // table alias -> real table name
	realNameAliases := make(map[string]string)
	pkTables := make([]int, 1)
	fromIdx := 0
	totalFltCnt := 0
	sortPriorityIdx := make([]int, 0, 1)
	realTableAliases := make([]string, len(joinTables))
	for idx := range joinTables {
		tInfo := &joinTables[idx]
		if tInfo.IsFrom {
			if fromIdx != 0 {
				return JoinerBase{}, fmt.Errorf("joiner can't use `from` more then once for tables %d and %d",
					fromIdx, idx)
			}
			fromIdx = idx
		} else if tInfo.IsPK && !slices.Contains(pkTables, idx) {
			pkTables = append(pkTables, idx) //nolint:makezero
		}

		realAlias := tInfo.Alias
		if realAlias == "" {
			for {
				aliasCnt++
				realAlias = fmt.Sprintf("T%d", aliasCnt)
				if _, ok := aliasPosMap[realAlias]; !ok {
					break
				}
			}
		} else if aliasIdx, ok := aliasPosMap[realAlias]; ok {
			return JoinerBase{}, fmt.Errorf("joiner can't use `alias` more then once for tables %d and %d", aliasIdx, idx)
		}
		realTableAliases[idx] = realAlias
		aliasNameMap[realAlias] = tInfo.TableInfo.SQLName
		realNameAliases[tInfo.TableInfo.SQLName] = realAlias
		totalFltCnt += len(tInfo.TableInfo.SelectIdxList)

		if tInfo.SortPriority != 0 {
			sortPriorityIdx = append(sortPriorityIdx, idx)
		}
	}
	if len(sortPriorityIdx) == 0 {
		sortPriorityIdx = append(sortPriorityIdx, fromIdx)
	} else {
		slices.SortStableFunc(sortPriorityIdx, func(a, b int) int {
			return joinTables[a].SortPriority - joinTables[b].SortPriority
		})
	}

	// Query build
	pos := 0
	var allFields struct_info.TableFields
	fldPosMap := make(map[string]int)
	var selSb strings.Builder
	selSb.Grow(totalFltCnt * 25)
	selSb.WriteString(defs.SQLSelect)
	for idx, tInfo := range joinTables {
		for _, fIdx := range tInfo.TableInfo.SelectIdxList {
			fldAlias := realTableAliases[idx] + defs.SQLDot + tInfo.TableInfo.Fields[fIdx].SQLName
			if _, ok := fldPosMap[fldAlias]; ok {
				return JoinerBase{}, fmt.Errorf("duplicate full field name `%s`", fldAlias)
			}
			fldPosMap[fldAlias] = pos
			allFields = append(allFields, struct_info.TableField{
				Index:        nil,
				Path:         nil,
				SQLName:      fldAlias,
				RefTable:     "",
				RefField:     "",
				SortPos:      0,
				IsPK:         false,
				CanSelect:    false,
				CanInsert:    false,
				CanUpdate:    false,
				SortBackward: false,
			})
			if pos > 0 {
				selSb.WriteString(defs.SQLCommaSpace)
			}
			pos++
			selSb.WriteString(fldAlias)
		}
	}

	selSb.WriteString(defs.SQLFrom)
	selSb.WriteString(joinTables[fromIdx].TableInfo.SQLName)
	selSb.WriteString(defs.SQLAs)
	selSb.WriteString(realTableAliases[fromIdx])
	defaultJoin := InnerJoin
	if joinTables[fromIdx].IsPointer {
		defaultJoin = OuterJoin
	}
	for idx, tInfo := range joinTables {
		if idx == fromIdx {
			continue
		}

		joinMode := tInfo.JoinModeVal
		if joinMode == DefaultJoin {
			joinMode = defaultJoin
		}
		selSb.WriteString("\n")
		selSb.WriteString(joinMode.SQLName())
		selSb.WriteString(defs.SQLSpace)
		selSb.WriteString(tInfo.TableInfo.SQLName)
		selSb.WriteString(defs.SQLAs)
		selSb.WriteString(realTableAliases[idx])
		selSb.WriteString(defs.SQLOn)
		if len(tInfo.TableInfo.RefIdxList) == 0 {
			selSb.WriteString(defs.SQLTrue)
		}
		for rPos, rIdx := range tInfo.TableInfo.RefIdxList {
			if rPos > 0 {
				selSb.WriteString(defs.SQLAnd)
			}
			selSb.WriteString(realTableAliases[idx])
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(tInfo.TableInfo.Fields[rIdx].SQLName)
			selSb.WriteString(defs.SQLEquals)
			refTable := tInfo.TableInfo.Fields[rIdx].RefTable
			if refAlias, exists := tInfo.RefAliasMap[refTable]; exists {
				refTable = refAlias
			} else {
				refTable = realNameAliases[refTable]
			}
			selSb.WriteString(refTable)
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(tInfo.TableInfo.Fields[rIdx].RefField)
		}
	}
	getManySQL := selSb.String()

	selSb.WriteString(defs.SQLWhere)

	pkTables[0] = fromIdx
	phNum := 0
	for _, pkIdx := range pkTables {
		pkTable := joinTables[pkIdx]
		for _, fldIdx := range pkTable.TableInfo.PKIdxList {
			if phNum > 0 {
				selSb.WriteString(defs.SQLAnd)
			}
			selSb.WriteString(realTableAliases[pkIdx])
			selSb.WriteString(defs.SQLDot)
			selSb.WriteString(pkTable.TableInfo.Fields[fldIdx].SQLName)
			selSb.WriteString(defs.SQLEquals)
			selSb.WriteString(d.Placeholder(phNum + 1))
			phNum++
		}
	}

	// sort
	var sbSort strings.Builder
	sbSort.WriteString(defs.SQLOrderBy)
	sortCnt := 0
	for _, tblIdx := range sortPriorityIdx {
		sortTbl := joinTables[tblIdx]
		for _, fldIdx := range sortTbl.TableInfo.SortIdxList {
			if sortCnt > 0 {
				sbSort.WriteString(defs.SQLCommaSpace)
			}
			fld := sortTbl.TableInfo.Fields[fldIdx]
			sbSort.WriteString(realTableAliases[tblIdx])
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
	return JoinerBase{
		dialect:    d,
		joinTables: joinTables,
		allFields:  allFields,
		fldPosMap:  fldPosMap,
		sql: joinerSQLs{
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
	TableInfo    *struct_info.TableInfo
	TableType    reflect.Type
	Index        []int
	IsPointer    bool
	IsPK         bool
	IsFrom       bool
	SortPriority int
	JoinModeVal  JoinMode
	RefAliasMap  map[string]string
	Alias        string
}

var knownJoinTables sync.Map

func collectJoinTables(t reflect.Type) (result JoinTables, err error) {
	if t.Kind() != reflect.Struct {
		return nil, errors.New("expects a struct's type argument for joining")
	}

	if v, ok := knownJoinTables.Load(t); ok {
		return v.(JoinTables), nil //nolint:errcheck
	}

	result = make([]JoinTable, 0, t.NumField())
	for idx := range t.NumField() {
		if !t.Field(idx).IsExported() {
			continue
		}
		if t.Field(idx).Anonymous {
			var subRes JoinTables
			if subRes, err = collectJoinTables(t.Field(idx).Type); err != nil {
				return nil, err
			}
			for _, sub := range subRes {
				sub.Index = append(t.Field(idx).Index, sub.Index...)
				result = append(result, sub)
			}

			continue
		}

		joinTableFlags, processable := ParseJoinTableFlags(t.Field(idx).Tag.Get(struct_info.TagName))
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
		if isPtr {
			joinMode = LeftJoin
		}
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

		joinTable := JoinTable{
			TableInfo:    tableInfo,
			TableType:    tableType,
			IsPointer:    isPtr,
			IsPK:         joinTableFlags.IsPK,
			IsFrom:       joinTableFlags.IsFrom,
			SortPriority: sortPriority,
			JoinModeVal:  joinMode,
			RefAliasMap:  aliasMap,
			Index:        []int{idx},
			Alias:        joinTableFlags.Alias,
		}

		result = append(result, joinTable)
	}

	knownJoinTables.Store(t, result)

	return result, nil
}
