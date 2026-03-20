package connections

import (
	"context"
	"sqlsharder/internal/repository"
	"sqlsharder/pkg/logger"
)

//Read shard connection details from the metadata DB and opens actual connections to each shard.
// open acutal db conn and store in map(conn pool in memory) [projectId][shardId]:[shard_connection]

type ConnectionManager struct {
	store         *ConnectionStore
	projectRepo   *repository.ProjectRepository
	shardRepo     *repository.ShardRepository
	shardConnRepo *repository.ShardConnectionRepository
}

//share a single instance across the system. by passing pointer

func NewConnectionManager(store *ConnectionStore, projectRepo *repository.ProjectRepository, shardRepo *repository.ShardRepository,
	shardConnRepo *repository.ShardConnectionRepository) *ConnectionManager {

	return &ConnectionManager{
		store:         store,
		projectRepo:   projectRepo,
		shardRepo:     shardRepo,
		shardConnRepo: shardConnRepo,
	}
}

func (m *ConnectionManager) CheckConnectionHealth(ctx context.Context, projectId string, shardId string) (bool, error) {

	conn, err := m.store.Get(projectId, shardId)
	if err != nil {
		return false, err
	}
	err = conn.Ping()
	if err != nil {
		return false, err
	}

	return true, nil
}

// func to initiate connections for all projects of all shards no startup
func (m *ConnectionManager) InitiateConnectionsAll(ctx context.Context) error {
	projects, err := m.projectRepo.ProjectList(ctx)
	if err != nil {
		logger.Logger.Warn("Failed to get projects", "error", err)
		return err
	}
	for _, project := range projects {
		shards, err := m.shardRepo.ShardList(ctx, project.ID)
		if err != nil {
			logger.Logger.Warn("Failed to get shards for project", "project_id", project.ID, "error", err)
			continue
		}
		for _, shard := range shards {
			shdconnInfo, err := m.shardConnRepo.GetConnectionByShardID(ctx, shard.ID)
			if err != nil {
				logger.Logger.Warn("Failed to get connection for shard", "shard_id", shard.ID, "error", err)
				return err
			}
			shdConnDsn := buildDsn(shdconnInfo)

			db, err := newConnectionSetup(ctx, shdConnDsn)
			if err != nil {
				logger.Logger.Warn("Failed to create connection for shard", "shard_id", shard.ID, "error", err)
				return err
			}
			m.store.Set(project.ID, shard.ID, db)
		}
	}
	logger.Logger.Info("Successfully initiated connections for all projects and shards")
	return nil
}

// func to initiate connections for a active project and shard
func (m *ConnectionManager) InitiateActiveProjectShardConnections(ctx context.Context) error {

	activeProject, err := m.projectRepo.GetActiveProjectId(ctx)
	if err != nil {
		logger.Logger.Warn("Failed to get active project", "error", err)
		return err
	}
	if activeProject == "" {
		logger.Logger.Warn("No active project found")
		return nil
	}
	shards, err := m.shardRepo.ShardList(ctx, activeProject)
	if err != nil {
		logger.Logger.Warn("Failed to get shards for project", "project_id", activeProject, "error", err)
		return err
	}

	for _, shard := range shards {
		shdConnInfo, err := m.shardConnRepo.GetConnectionByShardID(ctx, shard.ID)
		if err != nil {
			logger.Logger.Warn("Failed to get connection for shard", "shard_id", shard.ID, "error", err)
			continue
		}
		shdConnDsn := buildDsn(shdConnInfo)

		db, err := newConnectionSetup(ctx, shdConnDsn)
		if err != nil {
			logger.Logger.Warn("Failed to create connection for shard", "shard_id", shard.ID, "error", err)
			return err
		}
		m.store.Set(activeProject, shard.ID, db)
	}

	logger.Logger.Info("Successfully initiated connection for active project and shards")
	return nil
}
