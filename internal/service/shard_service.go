package service

import (
	"context"
	"database/sql"
	"errors"
	"sqlsharder/internal/repository"
)

type CreateShardRequest struct {
	ProjectID string `json:"project_id"`
}

type ShardService struct {
	repo *repository.ShardRepository
}

func NewShardService(repo *repository.ShardRepository) *ShardService {
	return &ShardService{repo: repo}
}

var (
	ErrShardNotFound      = errors.New("shard not found")
	ErrShardDeleteBlocked = errors.New("shard delete blocked")
)

func (s *ShardService) CreateShard(ctx context.Context, req *CreateShardRequest) (*repository.Shard, error) {
	if req.ProjectID == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.ShardAdd(ctx, req.ProjectID)
}

func (s *ShardService) GetShards(ctx context.Context, projectID string) ([]repository.Shard, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.ShardList(ctx, projectID)
}

func (s *ShardService) DeleteShard(ctx context.Context, shardID string) error {
	err := s.repo.ShardDelete(ctx, shardID)
	if err != nil {
		if errors.Is(err, repository.ErrShardDeleteBlocked) {
			return ErrShardDeleteBlocked
		}
		if errors.Is(err, sql.ErrNoRows) {
			return ErrShardNotFound
		}
		return err
	}
	return nil
}

func (s *ShardService) ActivateShard(ctx context.Context, shardID string) error {
	return s.repo.ShardActivate(ctx, shardID)
}

func (s *ShardService) DeactivateShard(ctx context.Context, shardID string) error {
	return s.repo.ShardDeactivate(ctx, shardID)
}

func (s *ShardService) GetShardStatus(ctx context.Context, shardID string) (string, error) {
	status, err := s.repo.FetchShardStatus(ctx, shardID)
	if err != nil {
		return "", ErrShardNotFound
	}
	return status, nil
}
