package struct_info

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/mirrorru/cruds/helpers"
)

type TableInfo struct {
	SQLName       string
	Fields        TableFields
	PKIdxList     []int
	InsertIdxList []int
	UpdateIdxList []int
	SelectIdxList []int
	SortIdxList   []int
	RefIdxList    []int
	FieldNameIdx  map[string]int
}

type SQLNamer interface {
	SQLName() string
}

func GetTableInfo(t reflect.Type) (*TableInfo, error) {
	fields, err := CollectTableFields(t)
	if err != nil {
		return nil, err
	}
	result := &TableInfo{
		SQLName:       getTableName(t),
		Fields:        fields,
		FieldNameIdx:  make(map[string]int, len(fields)),
		PKIdxList:     make([]int, 0, 2),
		InsertIdxList: make([]int, 0, len(fields)),
		UpdateIdxList: make([]int, 0, len(fields)),
		SelectIdxList: make([]int, 0, len(fields)),
	}
	for idx, field := range fields {
		if prevIdx, ok := result.FieldNameIdx[field.SQLName]; ok {
			return nil, fmt.Errorf("field `%s` is duplicated with indexes %d and %d ", field.SQLName, prevIdx, idx)
		}
		result.FieldNameIdx[field.SQLName] = idx
		if field.IsPK {
			result.PKIdxList = append(result.PKIdxList, idx)
		}
		if field.CanSelect {
			result.SelectIdxList = append(result.SelectIdxList, idx)
		}
		if field.CanInsert {
			result.InsertIdxList = append(result.InsertIdxList, idx)
		}
		if field.CanUpdate {
			result.UpdateIdxList = append(result.UpdateIdxList, idx)
		}
		if field.SortPos != 0 {
			result.SortIdxList = append(result.SortIdxList, idx)
		}
		if field.RefField != "" {
			result.RefIdxList = append(result.RefIdxList, idx)
		}
	}

	// sort result.SortIdxList
	slices.SortStableFunc(result.SortIdxList, func(a, b int) int {
		return fields[a].SortPos - fields[b].SortPos
	})

	return result, nil
}

func getTableName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr { //nolint:govet
		t = t.Elem()
	}
	zero := reflect.New(t).Interface()
	if namer, ok := zero.(SQLNamer); ok {
		return namer.SQLName()
	}

	return helpers.ToSnakeCase(t.Name())
}
