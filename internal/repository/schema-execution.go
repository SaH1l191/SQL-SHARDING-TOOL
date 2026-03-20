package repository

import (
	"context"
	"database/sql"
	"time"
)

type SchemaExecutionStatus struct {
	ID         string    `json:"id"`
	SchemaId   string    `json:"schema_id"` //project_versions_schema_id
	ShardId    string    `json:"shard_id"`
	State      string    `json:"state"`
	ErrMsg     string    `json:"err_msg"`
	ExecutedAt time.Time `json:"executed_at"`
}

type SchemaExecutionStatusRepository struct {
	db *sql.DB
}

func NewSchemaExecutionStatusRepository(db *sql.DB) *SchemaExecutionStatusRepository {
	return &SchemaExecutionStatusRepository{
		db: db,
	}
}

// creates a new shard execution record related to a shard in its project version schema
func (r *SchemaExecutionStatusRepository) CreateSchemaExecutionRecord(ctx context.Context,
	shdStatus SchemaExecutionStatus) error {

	query := `INSERT INTO schema_execution_status 
		(id, schema_id, shard_id, state, err_msg, executed_at) VALUES ($1, $2, $3, $4, $5, NOW())`

	_, err := r.db.ExecContext(ctx, query, shdStatus.ID, shdStatus.SchemaId, shdStatus.ShardId, shdStatus.State, shdStatus.ErrMsg)
	if err != nil {
		return err
	}
	return nil
}

//  1. Fetch shard connection
//  2. Execute DDL on shard DB
//  3. IF success:
//     call this function → state = applied
//     ELSE:
//     call this function → state = failed

// updates the execution status of a schema on a shard related to a project version schema
func (r *SchemaExecutionStatusRepository) ExecuteSchemaExecution(
	ctx context.Context,
	schemaID string,
	shardID string,
	state string,
	errorMessage *string,
) error {

	query := `UPDATE schema_execution_status SET state = $1, error_message = $2 , executed_at = $3 
	WHERE schema_id = $4 AND shard_id = $5`

	result, err := r.db.ExecContext(ctx, query, state, errorMessage, time.Now(), schemaID, shardID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// checks if all shards have applied a given schema
func (r *SchemaExecutionStatusRepository) ExecuteSchemaCheckAppliedAll(ctx context.Context, schemaId string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM schema_execution_status
		WHERE schema_id = $1 AND state != 'applied'`

	var count int
	err := r.db.QueryRowContext(ctx, query, schemaId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// fetches all schema execution statuses for a given shard
func (r *SchemaExecutionStatusRepository) ExecutionStatusFetchStatusByShardID(
	ctx context.Context,
	shardID string,
) ([]SchemaExecutionStatus, error) {

	query := `
		SELECT id, schema_id, shard_id, state, error_messgae, executed_at
		FROM schema_execution_status
		WHERE shard_id = $1`

	rows, err := r.db.QueryContext(ctx, query, shardID)
	if err != nil {
		return nil, err
	}

	var records []SchemaExecutionStatus

	for rows.Next() {
		var record SchemaExecutionStatus
		err := rows.Scan(
			&record.ID,
			&record.SchemaId,
			&record.ShardId,
			&record.State,
			&record.ErrMsg,
			&record.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// fetches all shard execution statuses for a given schema by its project version schema ID
func (r *SchemaExecutionStatusRepository) ExecutionRecordsFetchStatusAll(
	ctx context.Context, schemaID string) ([]SchemaExecutionStatus, error) {

	query := `
		SELECT id, schema_id, shard_id, state, error_message, executed_at
		FROM schema_execution_status
		WHERE schema_id = $1`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		schemaID,
	)
	if err != nil {
		return nil, err
	}
	var records []SchemaExecutionStatus
	for rows.Next() {
		var record SchemaExecutionStatus
		err = rows.Scan(
			&record.ID,
			&record.SchemaId,
			&record.ShardId,
			&record.State,
			&record.ErrMsg,
			&record.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// resets the state of a schema execution for a given shard
// & updates the error message to null & back from failed to pending
func (r *SchemaExecutionStatusRepository) ExecutionShardResetState(
	ctx context.Context, schemaId string, shardId string) error {
	query := `UPDATE schema_execution_status SET state = 'pending', executed_at = NULL, 
	error_message = NULL WHERE schema_id = $1 AND shard_id = $2`

	result, err := r.db.ExecContext(ctx, query, schemaId, shardId)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

//fetches the failed schema execution statuses for a given schema by its project version schema ID
func (r *SchemaExecutionStatusRepository) ExecutionRecordsFetchStatusFailed(
	ctx context.Context,
	schemaID string,
) ([]SchemaExecutionStatus, error) {

	query := `
		SELECT id, schema_id, shard_id, state, error_message, executed_at
		FROM schema_execution_status
		WHERE schema_id = $1 AND state = 'failed'`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		schemaID,
	)
	if err != nil {
		return nil, err
	}

	var records []SchemaExecutionStatus

	for rows.Next() {
		var record SchemaExecutionStatus

		err = rows.Scan(
			&record.ID,
			&record.SchemaId,
			&record.ShardId,
			&record.State,
			&record.ErrMsg,
			&record.ExecutedAt,
		)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil

}
