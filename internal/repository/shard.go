package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"github.com/google/uuid"
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
var ErrShardDeleteBlocked = errors.New("shard_delete_blocked")


func (s *ShardRepository) ShardAdd(ctx context.Context,projectID string) (*Shard,error) {
	
	shardIndexes ,err := s.FetchShardIndexes(ctx,projectID)
	if err != nil {
		return nil,err
	} 
	nextShardIndex  := getIndex(shardIndexes)

	
	shard := &Shard {
		ID:  uuid.NewString(),
		ProjectID: projectID,
		ShardIndex: nextShardIndex,
		Status: "inactive",
		CreatedAt: time.Now(),
	}
 
	query := `INSERT INTO shards (id, project_id, shard_index, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = s.db.ExecContext(ctx, query, shard.ID, shard.ProjectID, 
		shard.ShardIndex, shard.Status, shard.CreatedAt)
	if err != nil {
		return nil, err
	}
	return shard,nil
}

func (s *ShardRepository) ShardList(ctx context.Context,projectID string) ([]Shard,error) {
	query := `
		SELECT id, project_id, shard_index, status, created_at FROM shards
		WHERE project_id = $1
		ORDER BY shard_index
	`
	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shards := make([]Shard, 0)
	for rows.Next() {
		var shard Shard
		err := rows.Scan(
			&shard.ID,
			&shard.ProjectID,
			&shard.ShardIndex,
			&shard.Status,
			&shard.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		shards = append(shards, shard)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	} 
	return shards, nil
}

func (s *ShardRepository) ShardDeleteAll(ctx context.Context, projectID string) error {
	query := `
		DELETE FROM shards WHERE project_id = $1
	`
	result, err := s.db.ExecContext(
		ctx,
		query,
		projectID,
	)
	if err != nil {
		return err
	} 
	_, err = result.RowsAffected()
	if err != nil {
		return err
	} 
	return nil 
}

func (s *ShardRepository) ShardDelete(ctx context.Context, shardID string) error {
	var status string 
	err := s.db.QueryRowContext(
		ctx,
		`SELECT status FROM shards WHERE id = $1`,
		shardID,
	).Scan(&status)
	if err != nil {
		return err
	} 
	if status == "active" {
		return ErrShardDeleteBlocked
	} 
	_, err = s.db.ExecContext(
		ctx,
		`DELETE FROM shards WHERE id = $1`,
		shardID,
	) 
	return err
}

func (s *ShardRepository) ShardDeactivate(ctx context.Context, shardID string) error {
	query := `
		UPDATE shards SET status = 'inactive' WHERE id = $1
	` 
	_, err := s.db.ExecContext(
		ctx,
		query,
		shardID,
	) 
	if err != nil {
		return err
	} 
	return nil
}

func (s *ShardRepository) ShardActivate(ctx context.Context, shardID string) error {
	query := `
		UPDATE shards SET status = 'active' WHERE id = $1
	`
	_, err := s.db.ExecContext(
		ctx,
		query,
		shardID,
	) 
	if err != nil {
		return err
	} 
	return nil 
}

func (s *ShardRepository) FetchShardStatus(ctx context.Context, shardID string) (string, error) {
	query := `
		SELECT status FROM shards WHERE id = $1
	` 
	rows := s.db.QueryRowContext(
		ctx,
		query,
		shardID,
	) 
	var status string 
	err := rows.Scan(&status) 
	if err != nil {
		return "", err
	} 
	return status, nil
}


func (s *ShardRepository) FetchShardIndexes(ctx context.Context, projectID string) ([]int, error) {
	query := `
		SELECT shard_index FROM shards WHERE project_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexList []int

	for rows.Next() {
		var index int
		err := rows.Scan(&index)
		if err != nil {
			return nil, err
		} 
		indexList = append(indexList, index)
	} 
	if err := rows.Err(); err != nil {
		return nil, err
	} 
	return indexList, nil
}
func getIndex(arr []int) int {
	if len(arr) == 0 {
		return 0
	} 
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max + 1
}
