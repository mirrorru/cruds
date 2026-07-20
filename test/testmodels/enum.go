//go:build smoke || crudsgen || functional

package testmodels

import (
	"database/sql/driver"
	"fmt"
)

const (
	enumStrUnknown  = "unknown"
	enumStrLabelN_A = "Неизвестно"
)

func makeScanMap[E fmt.Stringer](validVals []E) map[string]E {
	result := make(map[string]E, len(validVals))
	for _, v := range validVals {
		result[v.String()] = v
	}
	return result
}

func scan[E interface {
	~int32
	OrNothing() E
}](value interface{}, scanMap map[string]E) (val E, err error) {
	if value == nil {
		val = E(0)
		return
	}

	ok := true
	switch v := value.(type) {
	case string:
		val, ok = scanMap[v]
	case E:
		val = v
	case int64:
		val = E(v)
	case int32:
		val = E(v)
	case int16:
		val = E(v)
	case int8:
		val = E(v)
	case int:
		val = E(v)
	case uint64:
		val = E(v)
	case uint32:
		val = E(v)
	case uint16:
		val = E(v)
	case uint8:
		val = E(v)
	case uint:
		val = E(v)
	case []byte:
		val, ok = scanMap[string(v)]
	case *string:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *E:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *int64:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *int32:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *int16:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *int8:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *int:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *uint64:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *uint32:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *uint16:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *uint8:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *uint:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	case *[]byte:
		if v != nil {
			return scan[E](*v, scanMap)
		}
		ok = false
	}

	if ok {
		checkedVal := val.OrNothing()
		if checkedVal != val {
			ok = false
			val = checkedVal
		}
	}

	if !ok {
		err = fmt.Errorf("invalid type %T", value)
	}

	return
}

type GenderType int32

const (
	GenderTypeUnknown GenderType = 0
	GenderTypeMale    GenderType = 1
	GenderTypeFemale  GenderType = 2
)

func (t GenderType) String() string {
	switch t {
	case GenderTypeMale:
		return "male"
	case GenderTypeFemale:
		return "female"
	default:
		return enumStrUnknown
	}
}

func (t GenderType) Label() string {
	switch t {
	case GenderTypeMale:
		return "Мужской"
	case GenderTypeFemale:
		return "Женский"
	default:
		return enumStrLabelN_A
	}
}

func (t GenderType) OrNothing() GenderType {
	switch t {
	case GenderTypeMale, GenderTypeFemale:
		return t
	default:
		return GenderTypeUnknown
	}
}

var allGenderType = []GenderType{GenderTypeMale, GenderTypeFemale}

func AllGenderType() []GenderType {
	return allGenderType
}

var genderTypeMap = makeScanMap[GenderType](AllGenderType())

func (x *GenderType) Scan(value interface{}) (err error) {
	*x, err = scan(value, genderTypeMap)
	return
}

func (x GenderType) Value() (driver.Value, error) {
	return x.String(), nil
}
