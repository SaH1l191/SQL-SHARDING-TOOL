package executor

import (
	"context"
	"database/sql"
	"sqlsharder/internal/connections"
	shardrouter "sqlsharder/internal/shardRouter"
	"strings"
)

type ExecutionResult struct {
	ShardID      string
	Columns      []string
	Rows         [][]any
	RowsAffected int64
	Err          error
}

type Executor struct {
	connStore *connections.ConnectionStore
}

func NewExecutor(store *connections.ConnectionStore) *Executor {
	return &Executor{
		connStore: store,
	}
}

func (e *Executor) Execute(
	ctx context.Context, projectID string,
	sqlText string, plan *shardrouter.RoutingPlan) ([]ExecutionResult, error) {

	if plan.Mode == shardrouter.RoutingModeRejected {
		return nil, plan.RejectError
	}

	results := make([]ExecutionResult, 0, len(plan.Targets))

	for _, target := range plan.Targets {
		db, err := e.connStore.Get(projectID, string(target.ShardID))
		if err != nil {
			results = append(results, ExecutionResult{
				ShardID: string(target.ShardID),
				Err:     err,
			})
			continue
		}
		result := executeOnShard(ctx, db, string(target.ShardID), sqlText)
		results = append(results, result)
	}
	return results, nil
}

func executeOnShard(ctx context.Context, db *sql.DB, shardID, sqlText string) ExecutionResult {
	trimmed := strings.TrimSpace(strings.ToUpper(sqlText))
	isSelect := strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH")

	if isSelect {
		rows, err := db.QueryContext(ctx, sqlText)
		if err != nil {
			return ExecutionResult{ShardID: shardID, Err: err}
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return ExecutionResult{ShardID: shardID, Err: err}
		}

		data := make([][]any, 0)
		for rows.Next() {
			values := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				return ExecutionResult{ShardID: shardID, Err: err}
			}
			row := make([]any, len(values))
			copy(row, values)
			data = append(data, row)
		}
		if err := rows.Err(); err != nil {
			return ExecutionResult{ShardID: shardID, Err: err}
		}
		return ExecutionResult{ShardID: shardID, Columns: cols, Rows: data}
	}

	// INSERT / UPDATE / DELETE / DDL
	res, err := db.ExecContext(ctx, sqlText)
	if err != nil {
		return ExecutionResult{ShardID: shardID, Err: err}
	}
	affected, _ := res.RowsAffected()
	return ExecutionResult{ShardID: shardID, RowsAffected: affected}
}