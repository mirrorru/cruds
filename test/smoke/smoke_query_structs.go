//go:build smoke

package smoke

// UserRow тестовая структура пользователя
type UserRow struct {
	ID   int    `tbl:"pk;auto"`
	Name string `tbl:"sort=1"`
	Age  int
}

// CommentRow тестовая структура комментария с ref на user_row
type CommentRow struct {
	ID     int `tbl:"pk;auto"`
	UserID int `tbl:"ref=user_row:id"`
	Text   string
}

// UserCommentsQuery запрос пользователя с комментариями
type UserCommentsQuery struct {
	User    UserRow    `tbl:"from"`
	Comment CommentRow `tbl:"join=left;alias=c1"`
}

// UserCommentsQueryPtr запрос с pointer T-полем
type UserCommentsQueryPtr struct {
	User    UserRow     `tbl:"from"`
	Comment *CommentRow `tbl:"join=left;alias=c1"`
}

// PostRow тестовая структура поста
type PostRow struct {
	ID     int `tbl:"pk;auto"`
	UserID int `tbl:"ref=user_row:id"`
	Title  string
}

// MultiJoinQuery запрос с несколькими JOIN'ами
type MultiJoinQuery struct {
	User    UserRow    `tbl:"from;alias=u1"`
	Comment CommentRow `tbl:"join=left;alias=c1"`
	Post    PostRow    `tbl:"join=inner;alias=p1"`
}

// ClientRow тестовая структура клиента с кастомными типами
type ClientRow struct {
	ID       int64          `tbl:"pk;auto"`
	Name     ClientName     `tbl:"sort=1"`
	Birthday ClientBirthday `tbl:"col=birthday"`
	Gender   GenderType
}

func (ClientRow) SQLName() string { return "clients" }

// OrderRow тестовая структура заказа
type OrderRow struct {
	ID       int64  `tbl:"pk;auto"`
	ClientID int64  `tbl:"ref=clients:id"`
	Item     string
	Qty      int
}

func (OrderRow) SQLName() string { return "orders" }

// ClientOrdersQuery запрос клиента с заказами (оба non-pointer)
type ClientOrdersQuery struct {
	Client ClientRow `tbl:"from"`
	Order  OrderRow  `tbl:"join=left;alias=o1"`
}

// ClientOrdersQueryPtrClient запрос с pointer Client
type ClientOrdersQueryPtrClient struct {
	Client *ClientRow `tbl:"from"`
	Order  OrderRow   `tbl:"join=left;alias=o1"`
}

// ClientOrdersQueryPtrOrder запрос с pointer Order
type ClientOrdersQueryPtrOrder struct {
	Client ClientRow  `tbl:"from"`
	Order  *OrderRow  `tbl:"join=left;alias=o1"`
}

// ClientOrdersQueryPtrBoth запрос с обоими pointer полями
type ClientOrdersQueryPtrBoth struct {
	Client *ClientRow `tbl:"from"`
	Order  *OrderRow  `tbl:"join=left;alias=o1"`
}
