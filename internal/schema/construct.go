package schema

import "sqlsharder/internal/repository"

func addColsToLogicalSchema(
	projectID string,
	schema *LogicalSchema,
	cols []repository.Column,
) {

	schema.ProjectId = projectID

	for _, col := range cols {
		ensureTable(schema, col.TableName)

		schema.Tables[col.TableName].Columns[col.ColumnName] = &Column{
			Name:     col.ColumnName,
			DataType: col.DataType,
			Nullable: col.IsNullable,
			IsPk:     col.IsPK,
		}
	}
}

func addFKsToLogicalSchema(
	schema *LogicalSchema,
	fks []repository.FkEdges,
) {

	for _, fk := range fks {
		ensureTable(schema, fk.ChildTable)
		key := Fkkey{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
		schema.Tables[fk.ChildTable].Fks[key] = &Fk{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
	}
}

func ensureTable(schema *LogicalSchema, tableName string) {
	if _, ok := schema.Tables[tableName]; !ok {
		schema.Tables[tableName] = &Table{
			Columns: make(map[string]*Column),
			Fks:     make(map[Fkkey]*Fk),
		}
	}
}
