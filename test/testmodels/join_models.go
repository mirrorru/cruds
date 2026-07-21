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

type JoinDefaultPointer struct {
	From JoinPtrFromRow `crud:"sort=10"`
	Left *JoinPtrLeftRow
}

type JoinSample struct {
	From  JoinFromRow  `crud:"sort=10"`
	Inner JoinInnerRow `crud:"sort=20"`
	Left  *JoinLeftRow `crud:"join=left;alias=LV"`
}

type JoinPtrFromRow struct {
	ID   int64 `crud:"pk;auto"`
	Name string
}

func (JoinPtrFromRow) SQLName() string { return "join_ptr_from" }

type JoinPtrLeftRow struct {
	ID    int64 `crud:"pk;auto"`
	RefID int64 `crud:"ref=join_ptr_from:id"`
	Value string
}

func (JoinPtrLeftRow) SQLName() string { return "join_ptr_left" }
