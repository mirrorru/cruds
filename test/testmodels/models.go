//go:build smoke || crudsgen || functional

package testmodels

type UserRow struct {
	ID   int    `crud:"pk;auto"`
	Name string `crud:"sort=1"`
	Age  int
}

type ProductRow struct {
	ID    int     `crud:"pk;auto"` //nolint:lll
	Name  string  `crud:"sort=1"`
	Price float64
	Stock int
}

func (ProductRow) SQLName() string {
	return "products"
}

type IdNameAgeRowFilled struct {
	ID   int    `crud:"pk"`
	Name string `crud:"sort=1"`
	Age  *int
}

func (IdNameAgeRowFilled) SQLName() string {
	return "id_name_age_filled"
}

type FuncRow struct {
	ID    int    `crud:"pk;auto"`
	Name  string `crud:"sort=1"`
	Value int
}
