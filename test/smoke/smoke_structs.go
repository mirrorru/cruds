//go:build smoke

package smoke

type IdNameAgeRow struct {
	ID   int    `tbl:"pk;auto"`
	Name string `tbl:"sort=1"`
	Age  int
}

type IdNameAgeRowFilled struct {
	ID   int    `tbl:"pk"`
	Name string `tbl:"sort=1"`
	Age  *int
}

func (IdNameAgeRowFilled) SQLName() string {
	return "id_name_age_filled"
}
