package repository

import (
	"context"
	"database/sql"
)

type ShardConnection struct {
	ShardId      string `json:"shard_id"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
type ShardConnectionRepository struct {
	db *sql.DB
}

func NewShardConnectionRepository(db *sql.DB) *ShardConnectionRepository {
	return &ShardConnectionRepository{db: db}
}

func (r *ShardConnectionRepository) ConnectionCreate(ctx context.Context, sc *ShardConnection) error {
	query := `INSERT INTO shard_connections (shard_id, host, port, database_name, username, password, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, sc.ShardId, sc.Host, sc.Port, sc.DatabaseName, sc.Username, sc.Password, sc.CreatedAt, sc.UpdatedAt)

	if err != nil {
		return err
	}
	return nil

}

func (r *ShardConnectionRepository) ConnectionRemove(ctx context.Context, shardId string) error {
	query := `DELETE FROM shard_connections where shard_id=$1`
	_, err := r.db.ExecContext(ctx, query, shardId)
	if err != nil {
		return err
	}
	return nil
}

func (r *ShardConnectionRepository) GetConnectionByShardID(ctx context.Context, shardId string) (ShardConnection, error) {
	query := `SELECT shard_id, host, port, database_name, username, password, created_at, updated_at FROM shard_connections WHERE shard_id = $1`
	var sc ShardConnection
	err := r.db.QueryRowContext(ctx, query, shardId).Scan(&sc.ShardId, &sc.Host, &sc.Port, &sc.DatabaseName, &sc.Username, &sc.Password, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		return ShardConnection{}, err
	}
	return sc, nil
}

func (c *ShardConnectionRepository) ConnectionUpdate(ctx context.Context, connInfo ShardConnection) error {
	query := `UPDATE shard_connections SET host = $1, port = $2, database_name = $3, username = $4, password = $5, updated_at = NOW() 
	WHERE shard_id = $6`

	_, err := c.db.ExecContext(
		ctx,
		query,
		connInfo.Host,
		connInfo.Port,
		connInfo.DatabaseName,
		connInfo.Username,
		connInfo.Password,
		connInfo.ShardId,
	)

	if err != nil {
		return err
	}

	return nil

}
