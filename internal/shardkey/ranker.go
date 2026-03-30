package shardkey

import (
	"fmt"
	"sort"
	"sqlsharder/internal/schema"
	"strings"
)

// logic :
// score =
//   - importance (who depends on me?)
//   - connectivity (how many tables?)
//   - ownership (am I a FK?)
//   - centrality (do I point to a root?)
//   - stability (am I PK?)
//   - bad distribution (text?)
//
// O(F + C)              // preprocessing
// + O(K * F)            // scoring (dominant)
// + O(K log K)          // sorting
// func RankCandidates(
// 	candidates CandidateSet,
// 	fanout map[ColumnReference]FanoutStats,
// 	cols []repository.Column,
// 	fks []repository.FkEdges,
// ) []ShardKeyDecision {

// 	// build lookup maps once — O(n) each, used O(1) per candidate
// 	isFKChild := make(map[ColumnReference]bool)
// 	for _, fk := range fks {
// 		isFKChild[ColumnReference{fk.ChildTable, fk.ChildColumn}] = true
// 	}

// 	isPK := make(map[ColumnReference]bool)
// 	for _, col := range cols {
// 		if col.IsPK {
// 			isPK[ColumnReference{col.TableName, col.ColumnName}] = true
// 		}
// 	}

// 	colDataType := make(map[ColumnReference]string)
// 	for _, col := range cols {
// 		colDataType[ColumnReference{col.TableName, col.ColumnName}] = col.DataType
// 	}

// 	//tableName : value is a slice of all FK edges for current table , not all tables
// 	fksByChildTable := make(map[string][]repository.FkEdges)
// 	for _, fk := range fks {
// 		fksByChildTable[fk.ChildTable] = append(fksByChildTable[fk.ChildTable], fk)
// 	}

// 	var decisions []ShardKeyDecision

// 	for tableName, tableCandidates := range candidates {

// 		// score every candidate in this table, collect into a slice
// 		type scored struct {
// 			col     ColumnReference
// 			score   int
// 			reasons []string
// 		}
// 		var ranked []scored

// 		tableFks := fksByChildTable[tableName]

// 		for _, ref := range tableCandidates {
// 			score := 1 // baseline
// 			var reasons []string
// 			reasons = append(reasons, "baseline (+1)")

// 			// incoming FK references — how many things depend on this column
// 			f := fanout[ref]
// 			if f.IncomingFkCount > 0 {
// 				v := f.IncomingFkCount * 10
// 				score += v
// 				reasons = append(reasons, fmt.Sprintf("referenced by %d FK (%+d)", f.IncomingFkCount, v))
// 			}

// 			// distinct referencing tables
// 			if f.ReferencingTableCount > 0 {
// 				v := f.ReferencingTableCount * 5
// 				score += v
// 				reasons = append(reasons, fmt.Sprintf("across %d tables (%+d)", f.ReferencingTableCount, v))
// 			}

// 			// column is itself a FK child (ownership / co-location signal)
// 			if isFKChild[ref] {
// 				score += 20
// 				reasons = append(reasons, "FK child column (+20)")

// 				// rootAffinityBonus: if the parent this column points to is
// 				// itself highly referenced, co-locating here is even better
// 				if bonus := rootAffinityBonus(ref, tableFks, fanout); bonus > 0 {
// 					score += bonus
// 					reasons = append(reasons, fmt.Sprintf("points to root table (%+d)", bonus))
// 				}
// 			}

// 			// primary key
// 			if isPK[ref] {
// 				score += 10
// 				reasons = append(reasons, "primary key (+10)")
// 			}

// 			// text/varchar columns have poor hash distribution — penalise
// 			// Strings hash poorly
// 			// Can cause uneven distribution
// 			dt := strings.ToLower(colDataType[ref])
// 			if dt == "text" || dt == "varchar" || strings.HasPrefix(dt, "varchar") {
// 				score -= 15
// 				reasons = append(reasons, "text column (-15)")
// 			}

// 			ranked = append(ranked, scored{ref, score, reasons})
// 		}

// 		// FIX: sort all candidates so the winner is always ranked[0].
// 		// Without this, equal-score columns produce different results on
// 		// every run because Go map iteration order is random.
// 		sort.SliceStable(ranked, func(i, j int) bool {
// 			if ranked[i].score != ranked[j].score {
// 				return ranked[i].score > ranked[j].score // highest score first
// 			}
// 			// deterministic tie-break: alphabetical by table then column name
// 			if ranked[i].col.TableName != ranked[j].col.TableName {
// 				return ranked[i].col.TableName < ranked[j].col.TableName
// 			}
// 			return ranked[i].col.ColumnName < ranked[j].col.ColumnName
// 		})

// 		//guard to ensure 0 index exists !
// 		if len(ranked) > 0 {
// 			best := ranked[0]
// 			decisions = append(decisions, ShardKeyDecision{
// 				Table:   tableName,
// 				Column:  best.col,
// 				Score:   best.score,
// 				Reasons: best.reasons,
// 			})
// 		}
// 	}

// 	return decisions
// }

// rootAffinityBonus checks if the column is a FK pointing to a "root table"
// (a table that many other tables reference). If so, sharding by this column
// keeps related rows from orders, payments, reviews etc. on the same shard.
//
// Example:
//
//	orders.user_id → users.id
//	users.id is referenced by 3 tables (orders, payments, reviews)
//	→ bonus = 3 * 5 = +15
//	→ total for orders.user_id: +20 (FK child) + +15 (root affinity) = +35
//
// O(F)
// func rootAffinityBonus(
// 	ref ColumnReference,
// 	fks []repository.FkEdges,
// 	fanout map[ColumnReference]FanoutStats,
// ) int {
// 	for _, fk := range fks {
// 		// find the FK edge where ref is the child column
// 		if fk.ChildTable != ref.TableName || fk.ChildColumn != ref.ColumnName {
// 			continue
// 		}
// 		// check how many things reference the parent column
// 		parent := ColumnReference{TableName: fk.ParentTable, ColumnName: fk.ParentColumn}
// 		if stats, ok := fanout[parent]; ok && stats.IncomingFkCount > 0 {
// 			return stats.IncomingFkCount * 5
// 		}
// 	}
// 	return 0
// }

// RankTableCandidates(tableName, candidateCols, fanout, logicalSchema)

func RankTableCandidates(
	tableName string,
	candidateCols []ColumnReference,
	fanout map[ColumnReference]FanoutStats,
	logicalSchema *schema.LogicalSchema) []RankedCandidate {

	var ranked []RankedCandidate
	table := logicalSchema.Tables[tableName]

	for _, col := range candidateCols {

		fkStats, ok := fanout[col]
		if !ok {
			fkStats = FanoutStats{}
		}

		logicalSchemaCol := logicalSchema.Tables[tableName].Columns

		score, reasons := scoreColumn(col, logicalSchemaCol, fkStats, table, fanout)
		ranked = append(ranked, RankedCandidate{
			Column:  col,
			Score:   score,
			Reasons: reasons,
		})

	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return tieBreak(ranked[i].Column, ranked[j].Column)
	})
	return ranked
}

func scoreColumn(col ColumnReference,
	logicalSchemaCol map[string]*schema.Column,
	fkStats FanoutStats,
	table *schema.Table,
	fanout map[ColumnReference]FanoutStats) (int, []string) {

	colName := col.ColumnName
	var reasons []string

	score := 1
	reasons = append(reasons, "base score")

	if logicalSchemaCol[colName].IsPk {
		score += 10
		reasons = append(reasons, "is a primary key")
	}
	if fkStats.IncomingFkCount > 0 {
		score = score + fkStats.IncomingFkCount*10
		reasons = append(reasons, fmt.Sprintf("referenced by %d foreign keys", fkStats.IncomingFkCount))
	}

	if fkStats.ReferencingTableCount > 0 {
		score = score + fkStats.IncomingFkCount*5
		reasons = append(reasons, fmt.Sprintf("references %d tables", fkStats.ReferencingTableCount))
	}
	if isForeignKey(colName, table) {
		score = score + 20
		reasons = append(reasons, "is a foreign key")
		if bonus, reason := rootAffinityBonus(colName, table, fanout); bonus > 0 {
			score += bonus
			reasons = append(reasons, reason)
		}
	}
	if isTextColumn(logicalSchemaCol[colName]) {
		score -= 15
		reasons = append(reasons, "is a text column")
	}
	return score, reasons
}

func isForeignKey(colName string, table *schema.Table) bool {
	for _, fk := range table.Fks {
		if fk.ChildColumn == colName { // if the column we’re testing is the “child” side of any FK
			return true
		}
	}
	return false
}

func isTextColumn(col *schema.Column) bool {
	switch strings.ToLower(col.DataType) {
	case "text", "varchar", "char", "character varying":
		return true
	}
	return false
}

// when scores equal -> sort by table name then column name
func tieBreak(a, b ColumnReference) bool {
	if a.TableName != b.TableName {
		return a.TableName < b.TableName
	}
	return a.ColumnName < b.ColumnName
}

// checks if current column's foreign key refering table is being referred by many other tables
// basically checks if the referred table (aka parent table) is popular among many tables
func rootAffinityBonus(colName string, logicalSchematable *schema.Table, fanout map[ColumnReference]FanoutStats) (int, string) {

	fks := logicalSchematable.Fks

	for _, fk := range fks {
		if fk.ChildColumn != colName {
			continue
		}

		//found the referred table (aka parent table)
		parent := ColumnReference{
			TableName:  fk.ParentTable,
			ColumnName: fk.ParentColumn,
		}
		fkStats, ok := fanout[parent]
		if !ok {
			continue
		}
		if fkStats.IncomingFkCount > 0 {
			bonusAwarded := fkStats.IncomingFkCount * 5
			return bonusAwarded, fmt.Sprintf("parent table is referenced by %d tables", fkStats.IncomingFkCount)
		}
	}
	return 0, ""
}
