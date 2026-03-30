package schema

import "sqlsharder/internal/repository"

func FlattenColumn(schema *LogicalSchema) []repository.Column {
	var cols []repository.Column
	for tableName, table := range schema.Tables {
		for _, column := range table.Columns {
			cols = append(cols, repository.Column{
				ProjectID:  schema.ProjectId,
				TableName:  tableName,
				ColumnName: column.Name,
				DataType:   column.DataType,
				IsNullable: column.Nullable,
				IsPK:       column.IsPk,
			})
		}
	}
	return cols
}

func FlattenFKEdges(schema *LogicalSchema) []repository.FkEdges {
	var fks []repository.FkEdges
	for tableName, table := range schema.Tables {
		for _, fk := range table.Fks {
			fks = append(fks, repository.FkEdges{
				ProjectId:    schema.ProjectId,
				ParentTable:  fk.ParentTable,
				ParentColumn: fk.ParentColumn,
				ChildTable:   tableName,
				ChildColumn:  fk.ChildColumn,
			})
		}
	}
	return fks
}
