package shardkey

import "sqlsharder/internal/repository"

func ComputeFanout(fks []repository.FkEdges) map[ColumnReference]FanoutStats {

	// build final stats from the sets
	stats := make(map[ColumnReference]FanoutStats)

	// The inner map[string]struct{} acts as a set of table names.
	childTableSets := make(map[ColumnReference]map[string]struct{})

	//0(n) using hashmap[hashmap []:[]]
	for _, fk := range fks {
		parent := ColumnReference{fk.ParentTable, fk.ParentColumn}

		// Update stats for the parent column
		s := stats[parent]
		s.IncomingFkCount++ // increment incoming foreign key count
		stats[parent] = s

		// Initialize the set if it doesn't exist
		if childTableSets[parent] == nil {
			childTableSets[parent] = make(map[string]struct{})
		}

		// Add the child table to the set (duplicate writes are ignored)
		childTableSets[parent][fk.ChildTable] = struct{}{}
	}

	for col, tableSet := range childTableSets {
		s := stats[col]
		s.ReferencingTableCount = len(tableSet) // number of unique child tables
		stats[col] = s
	}

	return stats
}

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
