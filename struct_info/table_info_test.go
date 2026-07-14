package struct_info_test

import (
	"quick-crud/struct_info"
	"quick-crud/test/samples"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTableInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		reflectType reflect.Type
		expected    struct_info.TableInfo
	}{
		{reflectType: reflect.TypeOf(samples.IdNameSmpl{}),
			expected: struct_info.TableInfo{
				SQLName: "id_name_smpl",
				Fields: struct_info.TableFields{
					{
						Index:     []int{0},
						Path:      []string{"ID"},
						SQLName:   "id",
						IsPK:      true,
						CanSelect: true,
						CanInsert: true,
					},
					{
						Index:     []int{1},
						Path:      []string{"Name"},
						SQLName:   "name",
						CanSelect: true,
						CanInsert: true,
						CanUpdate: true,
						SortPos:   1,
					}},
				PKIdxList:     []int{0},
				InsertIdxList: []int{0, 1},
				UpdateIdxList: []int{1},
				SelectIdxList: []int{0, 1},
				SortIdxList:   []int{1},
				FieldNameIdx:  map[string]int{"id": 0, "name": 1},
			},
		},
		{reflectType: reflect.TypeOf(samples.TwoKey{}),
			expected: struct_info.TableInfo{SQLName: "two_key", Fields: struct_info.TableFields{
				struct_info.TableField{Index: []int{0}, Path: []string{"Key1"}, SQLName: "key1", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{1}, Path: []string{"Key2"}, SQLName: "key2", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
			}, PKIdxList: []int{0, 1}, InsertIdxList: []int{0, 1}, UpdateIdxList: []int{}, SelectIdxList: []int{0, 1}, SortIdxList: []int(nil), RefIdxList: []int(nil), FieldNameIdx: map[string]int{"key1": 0, "key2": 1}}},
		{reflectType: reflect.TypeOf(samples.TwoFields{}),
			expected: struct_info.TableInfo{SQLName: "two_fields", Fields: struct_info.TableFields{
				struct_info.TableField{Index: []int{0}, Path: []string{"Fld1"}, SQLName: "field_one", RefTable: "", RefField: "", SortPos: 0, IsPK: false, CanSelect: true, CanInsert: true, CanUpdate: true, SortBackward: false},
				struct_info.TableField{Index: []int{1}, Path: []string{"Fld2"}, SQLName: "field_two", RefTable: "", RefField: "", SortPos: 0, IsPK: false, CanSelect: true, CanInsert: true, CanUpdate: true, SortBackward: false},
			}, PKIdxList: []int{}, InsertIdxList: []int{0, 1}, UpdateIdxList: []int{0, 1}, SelectIdxList: []int{0, 1}, SortIdxList: []int(nil), RefIdxList: []int(nil), FieldNameIdx: map[string]int{"field_one": 0, "field_two": 1}}},
		{reflectType: reflect.TypeOf(samples.TwoCombo1Smpl{}),
			expected: struct_info.TableInfo{SQLName: "sample_table", Fields: struct_info.TableFields{
				struct_info.TableField{Index: []int{0, 0}, Path: []string{"TwoKey", "Key1"}, SQLName: "key1", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{0, 1}, Path: []string{"TwoKey", "Key2"}, SQLName: "key2", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{1, 0}, Path: []string{"TwoFields", "Fld1"}, SQLName: "field_one", RefTable: "", RefField: "", SortPos: 0, IsPK: false, CanSelect: true, CanInsert: true, CanUpdate: true, SortBackward: false},
				struct_info.TableField{Index: []int{1, 1}, Path: []string{"TwoFields", "Fld2"}, SQLName: "field_two", RefTable: "", RefField: "", SortPos: 0, IsPK: false, CanSelect: true, CanInsert: true, CanUpdate: true, SortBackward: false},
			}, PKIdxList: []int{0, 1}, InsertIdxList: []int{0, 1, 2, 3}, UpdateIdxList: []int{2, 3}, SelectIdxList: []int{0, 1, 2, 3}, SortIdxList: []int(nil), RefIdxList: []int(nil), FieldNameIdx: map[string]int{"field_one": 2, "field_two": 3, "key1": 0, "key2": 1}}},
		{reflectType: reflect.TypeOf(samples.TwoCombo2Smpl{}),
			expected: struct_info.TableInfo{SQLName: "two_combo2_smpl", Fields: struct_info.TableFields{
				struct_info.TableField{Index: []int{0, 0}, Path: []string{"TwoKey", "Key1"}, SQLName: "key1", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{0, 1}, Path: []string{"TwoKey", "Key2"}, SQLName: "key2", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{1}, Path: []string{"TwoFlds"}, SQLName: "two_flds", RefTable: "", RefField: "", SortPos: 0, IsPK: false, CanSelect: true, CanInsert: true, CanUpdate: true, SortBackward: false},
			}, PKIdxList: []int{0, 1}, InsertIdxList: []int{0, 1, 2}, UpdateIdxList: []int{2}, SelectIdxList: []int{0, 1, 2}, SortIdxList: []int(nil), RefIdxList: []int(nil), FieldNameIdx: map[string]int{"key1": 0, "key2": 1, "two_flds": 2}}},
		{reflectType: reflect.TypeOf(samples.TwoCombo3Smpl{}),
			expected: struct_info.TableInfo{SQLName: "two_combo3_smpl", Fields: struct_info.TableFields{
				struct_info.TableField{Index: []int{0, 0}, Path: []string{"TwoKey", "Key1"}, SQLName: "key1", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{0, 1}, Path: []string{"TwoKey", "Key2"}, SQLName: "key2", RefTable: "", RefField: "", SortPos: 0, IsPK: true, CanSelect: true, CanInsert: true, CanUpdate: false, SortBackward: false},
				struct_info.TableField{Index: []int{1}, Path: []string{"TwoFlds"}, SQLName: "two_flds", RefTable: "", RefField: "", SortPos: 0, IsPK: false, CanSelect: true, CanInsert: true, CanUpdate: true, SortBackward: false},
			}, PKIdxList: []int{0, 1}, InsertIdxList: []int{0, 1, 2}, UpdateIdxList: []int{2}, SelectIdxList: []int{0, 1, 2}, SortIdxList: []int(nil), RefIdxList: []int(nil), FieldNameIdx: map[string]int{"key1": 0, "key2": 1, "two_flds": 2}}},
	}

	for _, tt := range tests {
		t.Run(tt.reflectType.Name(), func(t *testing.T) {
			got, err := struct_info.GetTableInfo(tt.reflectType)
			assert.NoError(t, err, tt.reflectType.Name())
			assert.Equal(t, tt.expected, got, tt.reflectType.Name())
		})
	}
}

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
