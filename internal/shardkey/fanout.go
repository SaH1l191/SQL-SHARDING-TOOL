package shardkey

import (
	"sqlsharder/internal/schema"
)

// func ComputeFanout(fks []repository.FkEdges) map[ColumnReference]FanoutStats {

// 	// build final stats from the sets
// 	stats := make(map[ColumnReference]FanoutStats)

// 	// The inner map[string]struct{} acts as a set of table names.
// 	childTableSets := make(map[ColumnReference]map[string]struct{})

// 	//0(n) using hashmap[hashmap []:[]]
// 	for _, fk := range fks {
// 		parent := ColumnReference{fk.ParentTable, fk.ParentColumn}

// 		// Update stats for the parent column
// 		s := stats[parent]
// 		s.IncomingFkCount++ // increment incoming foreign key count
// 		stats[parent] = s

// 		// Initialize the set if it doesn't exist
// 		if childTableSets[parent] == nil {
// 			childTableSets[parent] = make(map[string]struct{})
// 		}

// 		// Add the child table to the set (duplicate writes are ignored)
// 		childTableSets[parent][fk.ChildTable] = struct{}{}
// 	}

// 	for col, tableSet := range childTableSets {
// 		s := stats[col]
// 		s.ReferencingTableCount = len(tableSet) // number of unique child tables
// 		stats[col] = s
// 	}

// 	return stats
// }

// Input :
// fks := []repository.FkEdges{
//     {ChildTable: "orders", ChildColumn: "user_id", ParentTable: "users", ParentColumn: "id"},
//     {ChildTable: "payments", ChildColumn: "user_id", ParentTable: "users", ParentColumn: "id"},
//     {ChildTable: "reviews", ChildColumn: "user_id", ParentTable: "users", ParentColumn: "id"},
// }

// Initial state:

// stats = {}
// childTableSets = {}

// Iteration 1:

// fk = {ChildTable: "orders", ParentTable: "users", ParentColumn: "id"}
// parent = ColumnReference{TableName: "users", ColumnName: "id"}
// s := stats[parent] → s = {IncomingFkCount:0, ReferencingTableCount:0}
// Increment IncomingFkCount → s.IncomingFkCount = 1
// Update stats → stats[parent] = {IncomingFkCount:1, ReferencingTableCount:0}
// childTableSets[parent] is nil, initialize → childTableSets[parent] = {}
// Add "orders" → childTableSets[parent] = {"orders": struct{}{}}

// State after iteration 1:

// stats = {
//     {users, id}: {IncomingFkCount:1, ReferencingTableCount:0}
// }
// childTableSets = {
//     {users, id}: {"orders": struct{}{}}
// }

//v1

func ComputeFanout(logicalSchema *schema.LogicalSchema, candidateTableColumns CandidateSet) map[ColumnReference]FanoutStats {
	//output : map[ColumnReference]FanoutStats

	// 	type FanoutStats struct {
	// 	IncomingFkCount       int // how many foreign keys(col) reference this column
	// 	ReferencingTableCount int // referencing unique table count
	// }
	// type ColumnReference struct {
	// 	TableName  string
	// 	ColumnName string
	// }

	candidateColSet := make(map[ColumnReference]struct{})
	for tableName, cols := range candidateTableColumns {
		for _, col := range cols {
			candidateColSet[ColumnReference{
				TableName:  tableName,
				ColumnName: col.ColumnName,
			}] = struct{}{}
		}
	}

	fanout := make(map[ColumnReference]FanoutStats)
	refTables := make(map[ColumnReference]map[string]struct{}) // for each parent col, how many child tables reference it (set of strings)

	for _, table := range logicalSchema.Tables {
		for _, fkValue := range table.Fks {
			parentTable := fkValue.ParentTable
			parentColumn := fkValue.ParentColumn
			childTable := fkValue.ChildTable

			//uniquely identify the parent column
			parent := ColumnReference{
				TableName:  parentTable,
				ColumnName: parentColumn,
			}

			//only consider candidate columns, eliminated ones dont need fanout calculation
			if _, ok := candidateColSet[parent]; !ok {
				continue
			}

			prev := fanout[parent]
			prev.IncomingFkCount++
			fanout[parent] = prev

			if _, ok := refTables[parent]; !ok {
				refTables[parent] = make(map[string]struct{})
			}
			refTables[parent][childTable] = struct{}{}
		}
	}
	for parent, tableSet := range refTables {
		stats := fanout[parent]
		stats.ReferencingTableCount = len(tableSet)
		fanout[parent] = stats
	}
	return fanout
}
