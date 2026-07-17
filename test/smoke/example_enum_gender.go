package smoke

import (
	"database/sql/driver"
)

// GenderType — пол (int32, соответствует protobuf-enum).
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

// Scan implements the Scanner interface.
func (x *GenderType) Scan(value interface{}) (err error) {
	*x, err = scan(value, genderTypeMap)
	return
}

// Value implements the driver Valuer interface.
func (x GenderType) Value() (driver.Value, error) {
	return x.String(), nil
}
