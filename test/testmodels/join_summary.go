//go:build smoke || crudsgen || functional

package testmodels

type JoinRowFrom struct {
	ID       int64 `crud:"pk;auto;sort=1:desc"`
	Name     string
	Birthday ClientBirthday
	Gender   GenderType
}

func (JoinRowFrom) SQLName() string {
	return "table_from"
}

type JoinRowInnerJoin struct {
	ID        int64 `crud:"pk;auto"`
	RefID     int64 `crud:"ref=table_from:id;sort=1"`
	InnerName string
}

func (JoinRowInnerJoin) SQLName() string {
	return "table_inner"
}

type JoinRowLeftJoin struct {
	ID       int64 `crud:"pk;auto"`
	RefID    int64 `crud:"ref=table_from:id;sort=1:desc"`
	LeftName string
	Birthday ClientBirthday
	Gender   GenderType
}

func (JoinRowLeftJoin) SQLName() string {
	return "table_left"
}

type JoinRowAnonymousVal struct {
	InnerVal JoinRowInnerJoin `crud:"sort=20"`
	LeftVal  *JoinRowLeftJoin `crud:"join=left;alias=LV"`
}

type JoinSummary struct {
	From JoinRowFrom `crud:"sort=10"`
	JoinRowAnonymousVal
}
