package shardkey

import (
	"sqlsharder/internal/repository"
	"strings"
)

func shouldEliminate(col repository.Column) bool {

	if col.IsNullable {
		return true
	}

	dt := strings.ToLower(col.DataType)
	if dt == "bool" || dt == "boolean" {
		return true
	}

	name := strings.ToLower(col.ColumnName)
	if strings.HasPrefix(name, "is_") ||
		strings.Contains(name, "flag") ||
		strings.Contains(name, "status") {
		return true
	}

	switch name {
	case "created_at", "updated_at", "deleted_at", "version":
		return true
	}

	return false
}

// [tableName]:[[tName,Colname],[tName,Colname]...]
func ExtractCandidates(cols []repository.Column) CandidateSet {
	candidates := make(CandidateSet)
 
	for _, col := range cols {
		if shouldEliminate(col) {
			continue
		}
		candidates[col.TableName] = append(candidates[col.TableName], ColumnReference{
			TableName:  col.TableName,
			ColumnName: col.ColumnName,
		})
	}
 
	return candidates
}
