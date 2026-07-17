//go:build smoke || crudsgen || functional

package testmodels

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type ClientName string

type ClientBirthday struct {
	Date `crud:"col=birthday"`
}

func (cb *ClientBirthday) Scan(src interface{}) error {
	return cb.Date.Scan(src)
}

type Date time.Time

var (
	_ driver.Valuer = Date{}
	_ sql.Scanner   = (*Date)(nil)
)

func (d Date) String() string { return time.Time(d).String() }

func (d *Date) Parse(in string) error {
	t, err := time.Parse(time.DateOnly, in)
	*d = Date(t)
	return err
}

func (d Date) Value() (driver.Value, error) {
	return time.Time(d), nil
}

func (d *Date) Scan(src interface{}) error {
	switch v := src.(type) {
	case time.Time:
		*d = Date(v)
	case *time.Time:
		if v != nil {
			*d = Date(*v)
		}
	case string:
		t, err := time.Parse(time.DateOnly, v)
		if err != nil {
			return err
		}
		*d = Date(t)
	case nil:
		*d = Date{}
	default:
		return fmt.Errorf("не удалось преобразовать %T в Date", src)
	}
	return nil
}

func (d *Date) Set(t time.Time) {
	*d = Date(t)
}
