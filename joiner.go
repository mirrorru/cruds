package crudquick

import (
	"reflect"

	"github.com/mirrorru/crudquick/struct_info"
)

type Joiner[JT any] struct {
	// JT - joined tables
	tsType reflect.Type
}

func NewJoiner[TS any]() *Joiner[TS] {
	return nil
}

type JTInfo struct {
	tablesInfo []struct_info.TableInfo
}
