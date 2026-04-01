package executor

import (
	"context"
	"database/sql"
	"errors" 
	"sqlsharder/internal/connections"
	"sqlsharder/internal/repository"
	shardrouter "sqlsharder/internal/shardRouter"
	"sqlsharder/pkg/logger"
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

// executes projectSchema on all shards,updates shd record and all shard must be active and updates projectSchemaState accordingly
func ExecuteProjectSchema(ctx context.Context, projectId string, schemaRepo *repository.ProjectSchemaRepository,
	shardRepo *repository.ShardRepository, execRepo *repository.SchemaExecutionStatusRepository,
	executeDDL func(shardID string, ddl string) error,
) error {
	schema, err := schemaRepo.ProjectSchemaGetPending(ctx, projectId)
	if err != nil {
		return err
	}
	if err := schemaRepo.ProjectSchemaSetApplying(ctx, schema.ID); err != nil {
		return err
	}
	shds, err := shardRepo.ShardList(ctx, projectId)
	if err != nil {
		return err
	}
	for _, s := range shds {
		if err := execRepo.CreateSchemaExecutionRecord(ctx, repository.SchemaExecutionStatus{
			ID:       schema.ID,
			SchemaId: schema.ID,
			ShardId:  s.ID,
			State:    "pending",
		}); err != nil {
			logger.Logger.Error("failed to create schema execution record", "error", err)
			return err
		}

		if s.Status != "active" {
			msg := "shard inactive!"
			logger.Logger.Error("shard inactive", "shard_id", s.ID, "schema_id", schema.ID)
			_ = execRepo.ExecuteSchemaExecution(ctx, schema.ID, s.ID, "failed", &msg)
			_ = schemaRepo.ProjectSchemaUpdateSchemaState(ctx, schema.ID, "failed", &msg)
			return errors.New(msg)
		}
		if err := executeDDL(s.ID, schema.DDL_SQL); err != nil {
			msg := err.Error()
			logger.Logger.Error("failed to execute DDL on shard", "shard_id", s.ID, "schema_id", schema.ID, "error", msg)
			_ = execRepo.ExecuteSchemaExecution(ctx, schema.ID, s.ID, "failed", &msg)
			_ = schemaRepo.ProjectSchemaUpdateSchemaState(ctx, schema.ID, "failed", &msg)
			return err
		}
		logger.Logger.Info("DDL executed successfully on shard", "shard_id", s.ID, "schema_id", schema.ID)
		_ = execRepo.ExecuteSchemaExecution(ctx, schema.ID, s.ID, "applied", nil)
	}
	err = schemaRepo.ProjectSchemaUpdateSchemaState(ctx, schema.ID, "applied", nil)
	if err != nil {
		logger.Logger.Error("failed to update schema state", "error", err)
		return err
	}
	return nil
}

// sets failed shard id record states from failed back to pending & updats projcetSchema at last
func RetryFailedSchema(ctx context.Context, projectId string, schemaRepo *repository.ProjectSchemaRepository,
	execRepo *repository.SchemaExecutionStatusRepository) error {
	schema, err := schemaRepo.ProjectSchemaGetLatest(ctx, projectId)
	if err != nil {
		logger.Logger.Error("failed to get latest schema", "error", err)
		return err
	}
	if schema.State != "failed" {
		return errors.New("schema is not failed")
	}
	failedRecords, err := execRepo.ExecutionRecordsFetchStatusFailed(ctx, schema.ID)
	if err != nil {
		logger.Logger.Error("failed to fetch failed records", "error", err)
		return err
	}
	for _, r := range failedRecords {
		if err := execRepo.ExecutionShardResetState(ctx, schema.ID, r.ShardId); err != nil {
			logger.Logger.Error("failed to reset shard state", "error", err)
			return err
		}
	}
	if err := schemaRepo.ProjectSchemaUpdateSchemaState(ctx, schema.ID, "pending", nil); err != nil {
		logger.Logger.Error("failed to update schema state", "error", err)
		return err
	}
	return nil
}
