package repository

import (
	"context"
	"database/sql"
	"time"
)

type Shard struct {
	ID			string	`json:"id"`
	ProjectID	string	`json:"project_id"`
	ShardIndex	int		`json:"shard_index"`
	Status		string	`json:"status"`
	CreatedAt	time.Time `json:"created_at"`
}
 
type ShardRepository struct {
	db *sql.DB
}

func NewShardRepository(db *sql.DB) *ShardRepository {
	return &ShardRepository{
		db: db,
	}
}

func (s *ShardRepository) ShardAdd(ctx context.Context, shard *Shard) error {
	query := `INSERT INTO shards (id, project_id, shard_index, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.ExecContext(ctx, query, shard.ID, shard.ProjectID, shard.ShardIndex, shard.Status, shard.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}