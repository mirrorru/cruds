package smoke

type IdNameAgeRow struct {
	ID   int    `tbl:"pk"`
	Name string `tbl:"sort=1"`
	Age  *int
}
