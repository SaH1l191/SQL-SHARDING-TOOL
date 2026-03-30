package schema

type LogicalSchema struct {
	ProjectId string
	Tables    map[string]*Table
}

type Table struct {
	Columns map[string]*Column
	Fks     map[Fkkey]*Fk
}

type Column struct {
	Name     string
	DataType string
	Nullable bool
	IsPk     bool
}
type Fkkey struct {
	ChildColumn  string
	ParentTable  string
	ParentColumn string
}
type Fk struct {
	ChildTable   string
	ChildColumn  string
	ParentTable  string
	ParentColumn string
}

func NewLogicalSchema() *LogicalSchema {
	return &LogicalSchema{
		Tables: make(map[string]*Table),
	}
}

//3 tables:
// users
// orders
// payments

// orders.user_id → users.id
// payments.user_id → users.id
// payments.order_id → orders.id

// map[string]*Table{
//     "users":   &Table{...},
//     "orders":  &Table{...},
//     "payments": &Table{...},
// }

// ordersTable := &Table{
//     Columns: map[string]*Column{
//         "id":      {Name: "id"},
//         "user_id": {Name: "user_id"},
//     },
//     Fks: map[Fkkey]*Fk{
//         {
//             ChildColumn:  "user_id",
//             ParentTable:  "users",
//             ParentColumn: "id",
//         }: {
//             ChildTable:   "orders",
//             ChildColumn:  "user_id",
//             ParentTable:  "users",
//             ParentColumn: "id",
//         },
//     },
// }

// paymentsTable := &Table{
//     Fks: map[Fkkey]*Fk{
//         {
//             ChildColumn:  "user_id",
//             ParentTable:  "users",
//             ParentColumn: "id",
//         }: {
//             ChildTable:   "payments",
//             ChildColumn:  "user_id",
//             ParentTable:  "users",
//             ParentColumn: "id",
//         },
//         {
//             ChildColumn:  "order_id",
//             ParentTable:  "orders",
//             ParentColumn: "id",
//         }: {
//             ChildTable:   "payments",
//             ChildColumn:  "order_id",
//             ParentTable:  "orders",
//             ParentColumn: "id",
//         },
//     },
// }
