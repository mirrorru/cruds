package cruds

import (
	"strings"

	"github.com/mirrorru/cruds/struct_info"
)

type JoinTableTagFlags struct {
	IsFrom bool
	IsPK   bool
	Sort   string
	Alias  string
	Join   string
	Map    string
}

const (
	KeyTblFrom  = "from"
	KeyTblPK    = "pk"
	KeyTblOmit  = "omit"
	KeyTblSort  = "sort="
	KeyTblAlias = "alias="
	KeyTblJoin  = "join="
	KeyTblMap   = "map="
)

func parseJoinTableFlags(tag string) (result JoinTableTagFlags, ok bool) {
	keys := strings.Split(tag, struct_info.KeysSeparator)
	for _, key := range keys {
		switch {
		case key == KeyTblPK:
			result.IsPK = true
		case key == KeyTblFrom:
			result.IsFrom = true
		case key == KeyTblOmit:
			return result, false
		case struct_info.IsKey(KeyTblSort, key):
			result.Sort = key[len(KeyTblSort):]
		case struct_info.IsKey(KeyTblAlias, key):
			result.Alias = key[len(KeyTblAlias):]
		case struct_info.IsKey(KeyTblJoin, key):
			result.Join = key[len(KeyTblJoin):]
		case struct_info.IsKey(KeyTblMap, key):
			result.Map = key[len(KeyTblMap):]
		}
	}
	return result, true
}
