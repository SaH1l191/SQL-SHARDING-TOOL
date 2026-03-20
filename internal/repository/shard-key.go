package repository

import (
	"context"
	"database/sql"
	"time"
)

type ShardKeys struct {
	ProjectId        string    `json:"project_id"`
	TableName        string    `json:"table_name"`
	ShardKeyColumn   string    `json:"shard_key_column"`
	IsManualOverride bool      `json:"is_manual_override"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type ShardKeyRecord struct {
	TableName      string
	ShardKeyColumn string
	IsManual       bool
}

type ShardKeysRepository struct {
	db *sql.DB
}

func NewShardKeysRepository(db *sql.DB) *ShardKeysRepository {
	return &ShardKeysRepository{db: db}
}

func (s *ShardKeysRepository) FetchShardKeysByProjectID(
	ctx context.Context, projectID string) ([]ShardKeys, error) {

	query :=
		`SELECT project_id, table_name, shard_key_column, is_manual_override, updated_at
		FROM table_shard_keys
		WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []ShardKeys

	for rows.Next() {
		var key ShardKeys
		if err := rows.Scan(
			&key.ProjectId,
			&key.TableName,
			&key.ShardKeyColumn,
			&key.IsManualOverride,
			&key.UpdatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (s *ShardKeysRepository) ReplaceShardKeysForProject(ctx context.Context,
	projectID string,
	records []ShardKeyRecord,
) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteQuery :=
		`DELETE FROM table_shard_keys
		WHERE project_id = $1
		AND is_manual_override = FALSE`

	if _, err := tx.ExecContext(ctx, deleteQuery, projectID); err != nil {
		return err
	}
	//Insert new shard key OR
	// If already exists → update it
	//UPSERT (INSERT ... ON CONFLICT)
	insertQuery :=
		`INSERT INTO table_shard_keys
		(project_id, table_name, shard_key_column, is_manual_override, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (project_id, table_name)
		DO UPDATE SET
		shard_key_column = EXCLUDED.shard_key_column,
		is_manual_override = EXCLUDED.is_manual_override,
		updated_at = EXCLUDED.updated_at`

	for _, r := range records {
		if _, err := tx.ExecContext(
			ctx,
			insertQuery,
			projectID,
			r.TableName,
			r.ShardKeyColumn,
			r.IsManual,
			time.Now(),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}
