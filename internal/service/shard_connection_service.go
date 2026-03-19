package service

import (
	"context"
	"database/sql"
	"errors"
	"sqlsharder/internal/repository"
	"time"
)

type CreateShardConnectionRequest struct {
	ShardId      string `json:"shard_id"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type UpdateShardConnectionRequest struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type ShardConnectionService struct {
	repo *repository.ShardConnectionRepository
}

func NewShardConnectionService(repo *repository.ShardConnectionRepository) *ShardConnectionService {
	return &ShardConnectionService{repo: repo}
}

var (
	ErrShardConnectionNotFound = errors.New("shard connection not found")
)

func (s *ShardConnectionService) CreateShardConnection(ctx context.Context, req *CreateShardConnectionRequest) (*repository.ShardConnection, error) {
	if req.ShardId == "" || req.Host == "" || req.DatabaseName == "" || req.Username == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}
	if req.Port <= 0 {
		return nil, ErrInvalidInput
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	conn := &repository.ShardConnection{
		ShardId:      req.ShardId,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := s.repo.ConnectionCreate(ctx, conn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *ShardConnectionService) GetShardConnection(ctx context.Context, shardId string) (*repository.ShardConnection, error) {
	if shardId == "" {
		return nil, ErrInvalidInput
	}

	conn, err := s.repo.GetConnectionByShardID(ctx, shardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrShardConnectionNotFound
		}
		return nil, err
	}

	return &conn, nil
}

func (s *ShardConnectionService) UpdateShardConnection(ctx context.Context, shardId string, req *UpdateShardConnectionRequest) (*repository.ShardConnection, error) {
	if shardId == "" || req.Host == "" || req.DatabaseName == "" || req.Username == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}
	if req.Port <= 0 {
		return nil, ErrInvalidInput
	}

	existingConn, err := s.repo.GetConnectionByShardID(ctx, shardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrShardConnectionNotFound
		}
		return nil, err
	}

	updatedConn := repository.ShardConnection{
		ShardId:      shardId,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
		CreatedAt:    existingConn.CreatedAt,
		UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	err = s.repo.ConnectionUpdate(ctx, updatedConn)
	if err != nil {
		return nil, err
	}

	return &updatedConn, nil
}

func (s *ShardConnectionService) DeleteShardConnection(ctx context.Context, shardId string) error {
	if shardId == "" {
		return ErrInvalidInput
	}

	err := s.repo.ConnectionRemove(ctx, shardId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrShardConnectionNotFound
		}
		return err
	}
	return nil
}
