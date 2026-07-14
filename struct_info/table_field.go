package struct_info

import (
	"reflect"
	"slices"
)

type TableField struct {
	Index        []int
	Path         []string
	SQLName      string
	RefTable     string
	RefField     string
	SortPos      int
	IsPK         bool // keyPK
	CanSelect    bool
	CanInsert    bool
	CanUpdate    bool
	SortBackward bool
}
type TableFields []TableField

type fieldsIndexes struct {
	PKCols     []int
	SelectCols []int
	InsertCols []int
	UpdateCols []int
	SortCols   []int
	RefCols    []int
}

func (tfs TableFields) allIndexes() fieldsIndexes {
	result := fieldsIndexes{
		PKCols:     make([]int, 0, 2),
		SelectCols: make([]int, 0, len(tfs)),
		InsertCols: make([]int, 0, len(tfs)),
		UpdateCols: make([]int, 0, len(tfs)),
		// no SortCols preallocate
		// no RefCols preallocate
	}

	for idx, field := range tfs {
		if field.IsPK {
			result.PKCols = append(result.PKCols, idx)
		}
		if field.CanSelect {
			result.SelectCols = append(result.SelectCols, idx)
		}
		if field.CanInsert {
			result.InsertCols = append(result.InsertCols, idx)
		}
		if field.CanUpdate {
			result.UpdateCols = append(result.UpdateCols, idx)
		}
		if field.SortPos != 0 {
			result.SortCols = append(result.SortCols, idx)
		}
		if field.RefField != "" {
			result.RefCols = append(result.RefCols, idx)
		}
	}

	slices.SortStableFunc(result.RefCols, func(a, b int) int {
		return tfs[a].SortPos - tfs[b].SortPos
	})

	return result
}

func (tfs TableFields) ExtractRefs(src any, indexes []int) (refs []any) {
	rv := reflect.ValueOf(src).Elem()
	result := make([]any, len(indexes))
	for pos, idx := range indexes {
		fld := rv.FieldByIndex(tfs[idx].Index)
		result[pos] = fld.Addr().Interface()
	}

	return result
}
