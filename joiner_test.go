package crudquick_test

import (
	"fmt"
	"testing"

	"github.com/mirrorru/crudquick"
	"github.com/stretchr/testify/require"
)

type JoinRowFrom struct {
	ID   int64 `tbl:"pk;auto"`
	Name string
}

func (JoinRowFrom) SQLName() string {
	return "table_from"
}

type JoinRowInnerJoin struct {
	ID        int64 `tbl:"pk;auto"`
	RefID     int64 `tbl:"ref=table_from:id"`
	InnerName string
}

func (JoinRowInnerJoin) SQLName() string {
	return "table_inner"
}

type JoinRowLeftJoin struct {
	ID       int64  `tbl:"pk;auto"`
	RefID    int64  `tbl:"ref=table_from:id"`
	LeftName string `tbl:"sort=1"`
}

func (JoinRowLeftJoin) SQLName() string {
	return "table_left"
}

type JoinRowAnonymousVal struct {
	InnerVal JoinRowInnerJoin `tbl:"pk"`
	LeftVal  JoinRowLeftJoin  `tbl:"join=left;alias=LV"`
}

type JoinRowAnonymousRef struct {
	InnerRef *JoinRowInnerJoin `tbl:"map=table_from:LV"`
	LeftRef  *JoinRowLeftJoin  `tbl:"map=table_from:LV;sort=10"`
}

type JoinSummary struct {
	From JoinRowFrom
	JoinRowAnonymousVal
	JoinRowAnonymousRef
}

func TestNewJoiner(t *testing.T) {
	t.Parallel()
	join, err := crudquick.NewJoiner[JoinSummary](crudquick.SQLite)
	require.NoError(t, err)
	require.NotNil(t, join)
	for idx, table := range join.Tables() {
		fmt.Printf("%d:\n%#v\n", idx, table)
	}
	fmt.Println("GetOne:", join.SQLs().GetOneSQL)
	fmt.Println("Sort:", join.SQLs().SortSQL)
}
