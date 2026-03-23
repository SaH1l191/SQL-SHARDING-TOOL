package schema

import (
	"sqlsharder/internal/repository"
	"sqlsharder/pkg/logger"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// Flow :=

// ApplyDDL()
//    ↓
// ParseDDLToMetadata()
//    ↓
// parseDDLStatement() → AST
//    ↓
// extractFromCreateTable / extractFromAlterTable
//    ↓
// extractColumn + extractFKConstraint
//    ↓
// []columns + []fk_edges
//    ↓
// ReplaceExistingColumns()
// ReplaceFKEdgesForProject()

//for alter statements Flow :

//existing []Column (from DB)
//     +
// new []Column (from parsed DDL)
//     ↓
// merge by (tableName, columnName) key
//     ↓
// ReplaceExistingColumns with merged result

// ParseDDLToMetadata parses raw DDL text and returns flat slices of
// repository.Column and repository.FkEdges ready to be saved to the DB.
func ParseDDLToMetadata(projectId string, ddl string) ([]repository.Column, []repository.FkEdges, error) {
	ast, err := parseDDLStatement(ddl)
	if err != nil {
		return nil, nil, err
	}

	var columns []repository.Column
	var fkEdges []repository.FkEdges

	for _, rawStmt := range ast.Stmts {
		switch n := rawStmt.Stmt.Node.(type) {

		case *pg_query.Node_CreateStmt:
			cols, fks := extractFromCreateTable(projectId, n.CreateStmt)
			columns = append(columns, cols...)
			fkEdges = append(fkEdges, fks...)

		case *pg_query.Node_AlterTableStmt:
			cols, fks := extractFromAlterTable(projectId, n.AlterTableStmt)
			columns = append(columns, cols...)
			fkEdges = append(fkEdges, fks...)
		default:
			// ignore SET, COMMENT, etc.
			continue
		}
	}
	logger.Logger.Info("extracted ddl metadata", "columns", len(columns), "foreign_keys", len(fkEdges))
	return columns, fkEdges, nil
}

// extractFromCreateTable handles CREATE TABLE statements.
// Walks tableElts for ColumnDef (columns) and Constraint nodes (PKs, FKs).
func extractFromCreateTable(projectId string, stmt *pg_query.CreateStmt) ([]repository.Column, []repository.FkEdges) {
	tableName := stmt.Relation.Relname
	var cols []repository.Column
	var fks []repository.FkEdges

	for _, elt := range stmt.TableElts {
		switch e := elt.Node.(type) {

		case *pg_query.Node_ColumnDef:
			cols = append(cols, extractColumn(projectId, tableName, e.ColumnDef))

		case *pg_query.Node_Constraint:
			switch e.Constraint.Contype {

			case pg_query.ConstrType_CONSTR_FOREIGN:
				fks = append(fks, extractFKConstraint(projectId, tableName, e.Constraint)...)

			case pg_query.ConstrType_CONSTR_PRIMARY:
				// table-level PRIMARY KEY (id, ...) — mark matching columns
				for _, key := range e.Constraint.Keys {
					colName := getStringVal(key)
					for i := range cols {
						if cols[i].ColumnName == colName {
							cols[i].IsPK = true
						}
					}
				}
			}
		}
	}
	logger.Logger.Info("extracted create table", "table", tableName, "columns", len(cols), "foreign_keys", len(fks))
	return cols, fks
}

// extractFromAlterTable handles ALTER TABLE ADD COLUMN / ADD CONSTRAINT.
func extractFromAlterTable(projectId string, stmt *pg_query.AlterTableStmt) ([]repository.Column, []repository.FkEdges) {
	tableName := stmt.Relation.Relname
	var cols []repository.Column
	var fks []repository.FkEdges

	for _, cmd := range stmt.Cmds {
		c := cmd.Node.(*pg_query.Node_AlterTableCmd).AlterTableCmd

		switch c.Subtype {

		case pg_query.AlterTableType_AT_AddColumn:
			colDef := c.Def.Node.(*pg_query.Node_ColumnDef).ColumnDef
			cols = append(cols, extractColumn(projectId, tableName, colDef))

		case pg_query.AlterTableType_AT_AddConstraint:
			con := c.Def.Node.(*pg_query.Node_Constraint).Constraint
			if con.Contype == pg_query.ConstrType_CONSTR_FOREIGN {
				fks = append(fks, extractFKConstraint(projectId, tableName, con)...)
			}
		}
	}

	logger.Logger.Info("extracted alter table", "table", tableName, "columns", len(cols), "foreign_keys", len(fks))
	return cols, fks
}

// extractColumn converts a single ColumnDef AST node into a repository.Column.
func extractColumn(projectId, tableName string, colDef *pg_query.ColumnDef) repository.Column {
	nullable := true
	isPK := false

	for _, c := range colDef.Constraints {
		con := c.Node.(*pg_query.Node_Constraint).Constraint
		switch con.Contype {
		case pg_query.ConstrType_CONSTR_NOTNULL:
			nullable = false
		case pg_query.ConstrType_CONSTR_PRIMARY:
			isPK = true
			nullable = false // PKs are implicitly NOT NULL
		}
	}
	logger.Logger.Info("extracted column","table", tableName,"column", colDef.Colname)
	return repository.Column{
		ProjectID:  projectId,
		TableName:  tableName,
		ColumnName: colDef.Colname,
		DataType:   extractTypeName(colDef.TypeName),
		IsNullable: nullable,
		IsPK:       isPK,
	}
}

// extractFKConstraint converts a FOREIGN KEY constraint into one FkEdges
// entry per column pair (handles composite FKs).
func extractFKConstraint(projectId, tableName string, con *pg_query.Constraint) []repository.FkEdges {
	parentTable := con.Pktable.Relname
	var result []repository.FkEdges

	for i, fkAttr := range con.FkAttrs {
		result = append(result, repository.FkEdges{
			ProjectId:    projectId,
			ChildTable:   tableName,
			ChildColumn:  getStringVal(fkAttr),
			ParentTable:  parentTable,
			ParentColumn: getStringVal(con.PkAttrs[i]),
		})
	}
	logger.Logger.Info("extracted fk constraint","table", tableName,"fk_edges", len(result))
	return result
}

// extractTypeName pulls the type name string from a TypeName AST node.
// Uses the last name node — pg_query prefixes built-in types with "pg_catalog".
// e.g. names = ["pg_catalog", "int4"] → returns "int4"
func extractTypeName(t *pg_query.TypeName) string {
	if t == nil || len(t.Names) == 0 {
		return ""
	}

	return getStringVal(t.Names[len(t.Names)-1])
}

// getStringVal extracts the string value from a pg_query Node.
func getStringVal(n *pg_query.Node) string {
	if n == nil {
		return ""
	}
	if v, ok := n.Node.(*pg_query.Node_String_); ok {
		return v.String_.Sval
	}
	return ""
}
