package struct_info

import (
	"strconv"
	"strings"
)

// QueryTagFlags флаги тега tbl на полях Query-структуры.
// EN: QueryTagFlags flags of the tbl tag on Query struct fields.
type QueryTagFlags struct {
	IsFrom   bool
	JoinType string            // "left", "right", "inner", "" (auto)
	Alias    string            // SQL alias for this table
	RefMap   map[string]string // ref-table-name → alias
	IsPK     bool
	IsOmit   bool
	SortPos  int
}

// ParseQueryTag парсит тег tbl на полях Query-структуры.
// EN: ParseQueryTag parses the tbl tag on Query struct fields.
func ParseQueryTag(tag string) (QueryTagFlags, error) {
	result := QueryTagFlags{}
	if tag == "" {
		return result, nil
	}

	keys := strings.Split(tag, keySeparator)
	for _, key := range keys {
		switch {
		case key == "from":
			result.IsFrom = true
		case isKey("join=", key):
			result.JoinType = strings.ToLower(key[len("join="):])
		case isKey("alias=", key):
			result.Alias = key[len("alias="):]
		case isKey("map=", key):
			mapStr := key[len("map="):]
			result.RefMap = make(map[string]string)
			pairs := strings.Split(mapStr, inKeySeparator)
			for _, pair := range pairs {
				kv := strings.SplitN(pair, inKVSeparator, 2)
				if len(kv) == 2 {
					result.RefMap[kv[0]] = kv[1]
				}
			}
		case key == "pk":
			result.IsPK = true
		case key == "omit":
			result.IsOmit = true
		case isKey("sort=", key):
			pos, _ := strconv.Atoi(key[len("sort="):])
			result.SortPos = pos
		}
	}

	return result, nil
}
