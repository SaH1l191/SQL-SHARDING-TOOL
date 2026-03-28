package schema

func cloneLogicalSchema(base *LogicalSchema) *LogicalSchema {
	if base == nil {
		return NewLogicalSchema()
	}

	newSchema := NewLogicalSchema()
	newSchema.ProjectId = base.ProjectId

	for tableName, table := range base.Tables {
		newSchema.Tables[tableName] = cloneTable(table)
	}

	return newSchema
}

func cloneTable(t *Table) *Table {
	if t == nil {
		return &Table{
			Columns: make(map[string]*Column),
			Fks:     make(map[Fkkey]*Fk),
		}
	}

	newTable := &Table{
		Columns: make(map[string]*Column),
		Fks:     make(map[Fkkey]*Fk),
	}

	for colName, col := range t.Columns {
		newTable.Columns[colName] = &Column{
			Name:     col.Name,
			DataType: col.DataType,
			Nullable: col.Nullable,
			IsPk:     col.IsPk,
		}
	}

	for fkKey, fk := range t.Fks {
		newTable.Fks[fkKey] = &Fk{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
	}

	return newTable
}

func mergeTable(base *Table, delta *Table) *Table {
	merged := cloneTable(base)

	for colName, col := range delta.Columns {
		merged.Columns[colName] = &Column{
			Name:     col.Name,
			DataType: col.DataType,
			Nullable: col.Nullable,
			IsPk:     col.IsPk,
		}
	}

	for fkKey, fk := range delta.Fks {
		merged.Fks[fkKey] = &Fk{
			ChildColumn:  fk.ChildColumn,
			ParentTable:  fk.ParentTable,
			ParentColumn: fk.ParentColumn,
		}
	}

	return merged
}
