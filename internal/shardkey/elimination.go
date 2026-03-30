package shardkey

import (
	"sqlsharder/internal/schema"
	"strings"
)

// [tableName]:[[tName,Colname],[tName,Colname]...]
func ExtractCandidates(logicalSchema *schema.LogicalSchema) CandidateSet {
	candidates := make(CandidateSet)
	for tableName, table := range logicalSchema.Tables {
		for _, col := range table.Columns {
			if eliminated, _ := shouldEliminate(col); eliminated {
				continue
			}
			candidates[tableName] = append(candidates[tableName], ColumnReference{
				TableName:  tableName,
				ColumnName: col.Name,
			})
		}
	}
	return candidates
}

func shouldEliminate(col *schema.Column) (bool, string) {
	if col.Nullable {
		return true, "column is nullable"
	}
	name := strings.ToLower(col.Name)
	switch name {
	case "created_at", "updated_at", "version", "deleted_at":
		return true, "column is timestamp"
	}
	if strings.HasPrefix(name, "is_") ||
		strings.HasPrefix(name, "flag") ||
		strings.HasPrefix(name, "status") {
		return true, "column is boolean/flag/status"
	}
	return false, ""
}
