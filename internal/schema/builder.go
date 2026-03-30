package schema

import (
	"errors"
	"sqlsharder/internal/repository"
)

func BuildLogicalSchemaFromDDL(ddl string) (*LogicalSchema, error) {
	logicalSchema := NewLogicalSchema()
	ast, err := parseDDLStatement(ddl)
	if err != nil {
		return nil, err
	}
	err = extractSchemaFromAST(ast, logicalSchema)
	if err != nil {
		return nil, err
	}
	return logicalSchema, nil
}

func BuildLogicalSchemaFromDb(
	projectID string,
	columns []repository.Column,
	fkEdges []repository.FkEdges,
) (*LogicalSchema, error) {

	//working on duplicate schema to ensure consistency
	schema := NewLogicalSchema()
	schema.ProjectId = projectID

	addColsToLogicalSchema(projectID, schema, columns)
	addFKsToLogicalSchema(schema, fkEdges)

	return schema, nil

}

func MergeLogicalSchema(baseSchema *LogicalSchema, changes *LogicalSchema) (*LogicalSchema, error) {

	mergedSchema := cloneLogicalSchema(baseSchema) //immutability
	//mergedSchema , changes => append changes to mergedSchema
	for tableName, deltaTable := range changes.Tables {
		baseTable, exists := mergedSchema.Tables[tableName]
		if !exists {
			mergedSchema.Tables[tableName] = cloneTable(deltaTable)
			continue
		}
		mergedSchema.Tables[tableName] = mergeTable(baseTable, deltaTable)
	}
	return mergedSchema, nil
}

func FlattenLogicalSchema(schema *LogicalSchema) ([]repository.Column, []repository.FkEdges, error) {
	if schema == nil {
		return nil, nil, errors.New("nil columns or fk_edges")
	}
	columns := FlattenColumn(schema)
	fkEdges := FlattenFKEdges(schema)
	return columns, fkEdges, nil
}
