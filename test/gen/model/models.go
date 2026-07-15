package model

type UserRow struct {
	ID   int    `tbl:"pk;auto"`
	Name string `tbl:"sort=1"`
	Age  int
}

type ProductRow struct {
	ID    int     `tbl:"pk;auto"`
	Name  string  `tbl:"sort=1"`
	Price float64
	Stock int
}

func (ProductRow) SQLName() string {
	return "products"
}
