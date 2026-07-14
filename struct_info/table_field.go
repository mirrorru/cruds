package struct_info

import (
	"reflect"
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

func (tfs TableFields) ExtractRefs(src any, indexes []int) (refs []any) {
	rv := reflect.ValueOf(src).Elem()
	result := make([]any, len(indexes))
	for pos, idx := range indexes {
		fld := rv.FieldByIndex(tfs[idx].Index)
		result[pos] = fld.Addr().Interface()
	}

	return result
}

func (tfs TableFields) ExtractArgs(src any, indexes []int) []any {
	rv := reflect.ValueOf(src).Elem()
	result := make([]any, len(indexes))
	for pos, idx := range indexes {
		fld := rv.FieldByIndex(tfs[idx].Index)
		result[pos] = fld.Interface()
	}

	return result
}
