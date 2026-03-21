package shardkey

//ColumnRef{
//     Table:  "users",
//     Column: "id",
// }
//unique identifier for a column in entire schema
type ColumnReference struct {
	TableName  string
	ColumnName string
}

//tableName := []ColumnReference

//CandidateSet{
//     "users": [
//         {Table: "users", Column: "id"},
//         {Table: "users", Column: "email"},
//     ],
//     "orders": [
//         {Table: "orders", Column: "id"},
//         {Table: "orders", Column: "user_id"},
//     ],
// }

//list of valid shard key [tableName]:[[tName,Colname],[tName,Colname]...] candidates after filtering
type CandidateSet map[string][]ColumnReference

//Final answer for ONE table
// ShardKeyDecision{
//     Table: "users",
//     Column: ColumnReference{
//         Table:  "users",
//         Column: "id",
//     },
//     Score: 42,
// }
type FanoutStats struct {
	IncomingFkCount       int // how many foreign keys(col) reference this column
	ReferencingTableCount int // referencing unique table count
}

type ShardKeyDecision struct {
	Table  string
	Column ColumnReference
	Score  int
	Reasons []string
	// e.g. ["primary key (+10)", "referenced by 2 FK (+20)", ...]
}
