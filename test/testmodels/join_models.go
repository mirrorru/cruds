//go:build smoke || crudsgen || functional

package testmodels

type JoinFromRow struct {
	ID   int64 `crud:"pk;auto;sort=1:desc"`
	Name string
}

func (JoinFromRow) SQLName() string { return "join_from" }

type JoinInnerRow struct {
	ID       int64 `crud:"pk;auto"`
	RefID    int64 `crud:"ref=join_from:id;sort=1"`
	InnerVal string
}

func (JoinInnerRow) SQLName() string { return "join_inner" }

type JoinLeftRow struct {
	ID      int64 `crud:"pk;auto"`
	RefID   int64 `crud:"ref=join_from:id;sort=1:desc"`
	LeftVal string
}

func (JoinLeftRow) SQLName() string { return "join_left" }

type JoinSample struct {
	From  JoinFromRow  `crud:"sort=10"`
	Inner JoinInnerRow `crud:"sort=20"`
	Left  *JoinLeftRow `crud:"join=left;alias=LV"`
}
