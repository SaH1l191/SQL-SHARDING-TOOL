package schema 

// flow :=
//  1. -> convert input ddl
//  2. -> logical schema
//  3. -> convert to col/rows/table
//  4. -> save to DB (metadata/base schema to compare with )


type LogicalSchema struct {
	ProjectId string
	Tables    map[string]*Table
}

// all columns in a table
type Table struct {
	Columns map[string]*Column
	Fks     map[FkKey]*Fk
}

// column infomation
type Column struct {
	Name         string
	DataType     string
	Nullable     bool
	IsPrimaryKey bool
}

// describes what the relationship is.
type Fk struct {
	ChildTable   string
	ChildColumn  string
	ParentTable  string
	ParentColumn string
}

// describes how to uniquely identify it inside a table.
type FkKey struct {
	ChildColumn  string
	ParentTable  string
	ParentColumn string
}


func NewLogicalSchema() *LogicalSchema {
	return &LogicalSchema{
		Tables: make(map[string]*Table),
	}
}









