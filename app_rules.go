package main

import (
	"context"
	"database/sql"
	"sqlsharder/pkg/logger"
	"strings"
)

func (a *App) checkAllProjectsInactive() (bool, error) {
	projects, err := a.GetProjects()
	if err != nil {
		return false, err
	}
	for _, project := range projects {
		if project.Status == "active" {
			return false, nil
		}
	}
	return true, nil
}

func (a *App) checkAllShardsActive(projectID string) (bool, error) {
	shards, err := a.GetShards(projectID)
	if err != nil {
		return false, err
	}
	if len(shards) == 0 {
		return false, nil
	}
	for _, shard := range shards {
		if shard.Status != "active" {
			return false, nil
		}
	}
	return true, nil
}

// all schema := applied status
func (a *App) checkAllSchemaApplied(projectID string) (bool, error) {
	shards, err := a.ShardRepo.ShardList(a.ctx, projectID)
	if err != nil {
		return false, err
	}
	for _, shard := range shards {
		statuses, err := a.SchemaExecutionStatusRepo.ExecutionStatusFetchStatusByShardID(a.ctx, shard.ID)
		if err != nil {
			return false, err
		}
		for _, status := range statuses {
			if status.State != "applied" {
				return false, nil
			}
		}
	}
	return true, nil
}

// returns true if a live DB connection exists for the
func (a *App) checkIfShardConnected(projectID, shardID string) (bool, error) {
	return a.connManager.CheckConnectionHealth(a.ctx, projectID, shardID)
}

// A shard must be deactivated before it can be deleted.
func (a *App) checkIfShardInactive(shardID string) (bool, error) {
	status, err := a.ShardRepo.FetchShardStatus(a.ctx, shardID)
	if err != nil {
		return false, err
	}
	return status != "active", nil
}

// A schema may only be committed while the project is inactive.
func (a *App) checkIfProjectInactive(projectID string) (bool, error) {
	status, err := a.ProjectRepo.FetchProjectStatus(a.ctx, projectID)
	if err != nil {
		return false, err
	}
	return status == "inactive", nil
}

func (a *App) checkIfSchemaDraft(schemaID string) (bool, error) {
	state, err := a.ProjectSchemaRepo.ProjectSchemaGetState(a.ctx, schemaID)
	if err != nil {
		return false, err
	}
	return state == "draft", nil
}

func (a *App) checkIfSchemaInFlight(projectID string) (bool, error) {
	schemas, err := a.ProjectSchemaRepo.ProjectSchemaFetchHistory(a.ctx, projectID)
	if err != nil {
		return false, err
	}
	for _, schema := range schemas {
		if schema.State == "pending" || schema.State == "applying" {
			return true, nil
		}
	}
	return false, nil
}

func (a *App) checkIfDDLDestructive(projectID, ddlSQL string) (bool, error) {
	appliedSchema, err := a.ProjectSchemaRepo.ProjectSchemaGetApplied(a.ctx, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // no applied schema → nothing to protect
		}
		return false, err
	}
	if appliedSchema == nil {
		return false, nil
	}

	lowered := strings.ToLower(ddlSQL)
	destructive := strings.Contains(lowered, "drop table") ||
		strings.Contains(lowered, "drop column") ||
		strings.Contains(lowered, "truncate") ||
		(strings.Contains(lowered, "alter table") && strings.Contains(lowered, " drop "))

	return destructive, nil
}

// DML (INSERT / UPDATE / DELETE / SELECT / MERGE) is not allowed in the schema editor.
func (a *App) checkIfOnlyDDL(ddlSQL string) bool {
	lowered := strings.ToLower(ddlSQL)
	disallowed := []string{
		"insert ",
		"update ",
		"delete ",
		"select ",
		"merge ",
	}
	for _, kw := range disallowed {
		if strings.Contains(lowered, kw) {
			return false
		}
	}
	return true
}

func (a *App) checkShardHealth(ctx context.Context, projectID string, shardID string) (bool, error) {
	return a.connManager.CheckConnectionHealth(ctx, projectID, shardID)
}

// checkAllShards pings every shard of the active project and deactivates any
// that fail. Designed to be called from a background goroutine / ticker.
func (a *App) checkAllShards(ctx context.Context) {
	projectID, err := a.ProjectRepo.GetActiveProjectId(ctx)
	if err != nil {
		logger.Logger.Error("checkAllShards: failed to fetch active project", "error", err)
		return
	}
	shards, err := a.ShardRepo.ShardList(ctx, projectID)
	if err != nil {
		logger.Logger.Error("checkAllShards: failed to list shards", "error", err)
		return
	}
	for _, shard := range shards {
		healthy, err := a.connManager.CheckConnectionHealth(ctx, projectID, shard.ID)
		if err != nil || !healthy {
			_ = a.DeactivateShard(shard.ID)
			logger.Logger.Warn("Shard became unhealthy — deactivated",
				"project_id", projectID,
				"shard_id", shard.ID,
			)
		}
	}
}
