package schema

import "sqlsharder/internal/repository"

func FlattenColumn(schema *LogicalSchema) []repository.Column {

	var result []repository.Column

	for tableName, table := range schema.Tables {
		for _, column := range table.Columns {

			result = append(result, repository.Column{
				ProjectID:    schema.ProjectId,
				TableName:    tableName,
				ColumnName:   column.Name,
				DataType:     column.DataType,
				IsNullable:   column.Nullable,
				IsPK:         column.IsPk,
			})
		}
	}

	return result
}

func FlattenFKEdges(schema *LogicalSchema) []repository.FkEdges {

	var result []repository.FkEdges

	for tableName, table := range schema.Tables {
		for _, fk := range table.Fks {

			result = append(result, repository.FkEdges{
				ProjectId:    schema.ProjectId,
				ParentTable:  fk.ParentTable,
				ParentColumn: fk.ParentColumn,
				ChildTable:   tableName,
				ChildColumn:  fk.ChildColumn,
			})
		}
	}

	return result
}
