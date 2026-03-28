package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sqlsharder/internal/config"
	"sqlsharder/internal/connections"
	"sqlsharder/internal/executor"
	"sqlsharder/internal/loader"
	"sqlsharder/internal/repository"
	"sqlsharder/internal/schema"
	shardrouter "sqlsharder/internal/shardRouter"
	"sqlsharder/internal/shardkey"
	"sqlsharder/pkg/logger"
	"time"
)

type WailsConfig struct {
	WindowTitle  string
	WindowWidth  int
	WindowHeight int
	DevMode      bool
}

func DefaultWailsConfig() WailsConfig {
	return WailsConfig{
		WindowTitle:  "SQLSharder",
		WindowWidth:  1280,
		WindowHeight: 800,
		DevMode:      false,
	}
}

type App struct {
	ctx     context.Context
	cancel  context.CancelFunc
	db      *sql.DB
	server  *http.Server
	emitter *logger.LogEmitter

	connStore   *connections.ConnectionStore
	connManager *connections.ConnectionManager

	schemaService *schema.SchemaService
	shardRouter   *shardrouter.RouterService
	wailsCfg      WailsConfig

	ProjectRepo               *repository.ProjectRepository
	ShardRepo                 *repository.ShardRepository
	ProjectSchemaRepo         *repository.ProjectSchemaRepository
	ShardConnectionRepo       *repository.ShardConnectionRepository
	SchemaExecutionStatusRepo *repository.SchemaExecutionStatusRepository
	ShardKeysRepo             *repository.ShardKeysRepository
	FkEdgesRepo               *repository.FKEdgesRepository
	InferenceService          *shardkey.InferenceService
	ExecutorService           *executor.Executor
}

func NewApp(cfg WailsConfig) *App {
	return &App{wailsCfg: cfg}
}

// startup is called by Wails — must return quickly, never block.
func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	if err := a.init(); err != nil {
		logger.Logger.Error("App startup failed", "error", err)
	}
}

func (a *App) init() error {
	if err := loader.Loadservices(); err != nil {
		return fmt.Errorf("load services: %w", err)
	}

	var err error
	a.db, err = loader.LoadApplicationDatabase() //appDbConn stored in config
	if err != nil {
		return fmt.Errorf("load database: %w", err)
	}
	config.ApplicationDatabaseConnection.ConnInstance = a.db

	logger.Logger.Info("Database connected",
		"host", config.AppDBCreds.DB_HOST,
		"db", config.AppDBCreds.DB_NAME,
		"port", config.AppDBCreds.DB_PORT,
	)

	a.buildRepos()

	if err := a.connManager.InitiateConnectionsAll(a.ctx); err != nil {
		logger.Logger.Warn("Could not initiate shard connections", "error", err)
	}

	go func() {
		logger.Logger.Info("HTTP server starting", "addr", ":8080")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("HTTP server error", "error", err)
		}
	}()
	return nil
}
func (a *App) buildRepos() {
	projectRepo := repository.NewProjectRepository(a.db)
	shardRepo := repository.NewShardRepository(a.db)
	projectSchemaRepo := repository.NewProjectSchemaRepository(a.db)
	shardConnRepo := repository.NewShardConnectionRepository(a.db)
	schemaExecRepo := repository.NewSchemaExecutionStatusRepository(a.db)
	shardKeysRepo := repository.NewShardKeysRepository(a.db)
	fkEdgesRepo := repository.NewFKEdgesRepository(a.db)
	columnsRepo := repository.NewColumnRepository(a.db)
	a.ProjectRepo = projectRepo
	a.ShardRepo = shardRepo
	a.ProjectSchemaRepo = projectSchemaRepo
	a.ShardConnectionRepo = shardConnRepo
	a.SchemaExecutionStatusRepo = schemaExecRepo
	a.ShardKeysRepo = shardKeysRepo
	a.ExecutorService = executor.NewExecutor(a.connStore)
	a.connStore = connections.NewConnectionStore()
	a.connManager = connections.NewConnectionManager(
		a.connStore, projectRepo, shardRepo, shardConnRepo,
	)
	a.emitter = logger.NewLogEmitter(a.ctx)
	inferenceService := shardkey.NewInferenceService(columnsRepo, fkEdgesRepo, shardKeysRepo)
	a.InferenceService = inferenceService
	a.schemaService = schema.NewSchemaService(columnsRepo, fkEdgesRepo, inferenceService)
	a.ExecutorService = executor.NewExecutor(a.connStore)
	a.server = &http.Server{Addr: ":8080"}
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	a.cleanup()
}

func (a *App) cleanup() error {
	if a.server != nil {
		_ = a.server.Shutdown(a.ctx)
	}
	if a.db != nil {
		_ = a.db.Close()
	}
	logger.Logger.Info("Shutdown complete")
	return nil
}

func (a *App) GetWailsConfig() WailsConfig { return a.wailsCfg }

func (a *App) GetProjects() ([]*repository.Project, error) {
	result, err := a.ProjectRepo.ProjectList(a.ctx)
	if err != nil {
		logger.Logger.Error("Error while fetching projects", "error", err)
		a.emitter.Error("Projects listing failed", "application - ListProjects", map[string]string{
			"error": err.Error(),
		})
		return nil, err
	}
	return result, nil
}

func (a *App) CreateProject(name, description string) (*repository.Project, error) {
	result, err := a.ProjectRepo.ProjectAdd(a.ctx, name, description)
	if err != nil {
		logger.Logger.Error("Failed to create project", "project_name", name, "error", err)
		a.emitter.Error("Failed to create project", "create_project", map[string]string{
			"project_name": name,
			"error":        err.Error(),
		})
		return nil, err
	}
	logger.Logger.Info("Project created successfully", "project_name", name)
	a.emitter.Info("Project created successfully", "create_project", map[string]string{
		"project_name": name,
	})
	return result, nil
}

func (a *App) DeleteProject(id string) error {
	err := a.ProjectRepo.ProjectRemove(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Failed to delete project", "project_id", id, "error", err)
		a.emitter.Error("Failed to delete project", "delete_project", map[string]string{
			"project_id": id,
			"error":      err.Error(),
		})
		return err
	}
	logger.Logger.Info("Project deleted successfully", "project_id", id)
	a.emitter.Info("Project deleted successfully", "delete_project", map[string]string{
		"project_id": id,
	})
	return nil
}

func (a *App) GetProjectById(id string) (repository.Project, error) {
	project, err := a.ProjectRepo.GetProjectById(a.ctx, id)
	if err != nil {
		logger.Logger.Error("Failed to get project by ID", "project_id", id, "error", err)
		a.emitter.Error("Failed to get project by ID", "get_project_by_id", map[string]string{
			"project_id": id,
			"error":      err.Error(),
		})
		return repository.Project{}, err
	}
	return project, nil
}

func (a *App) AddShard(projectId string) (*repository.Shard, error) {
	result, err := a.ShardRepo.ShardAdd(a.ctx, projectId)
	if err != nil {
		logger.Logger.Error("Failed to add shard", "project_id", projectId, "error", err)
		a.emitter.Error("Failed to add shard", "add_shard", map[string]string{
			"project_id": projectId,
			"error":      err.Error(),
		})
		return nil, err
	}
	logger.Logger.Info("Shard added successfully", "project_id", projectId)
	a.emitter.Info("Shard added successfully", "add_shard", map[string]string{
		"project_id": projectId,
	})
	return result, nil
}

func (a *App) ActivateProject(projectId string) error {
	//all other projc must be inactive
	otherProjectsInactive, err := a.checkAllProjectsInactive()
	if err != nil {
		logger.Logger.Error("Failed to check status of projects for projct activation", "project_id", projectId, "error", err)
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectId,
			"error":      "unable to fetch projects for activation constratints",
		})
		return err
	}
	if otherProjectsInactive == false {
		logger.Logger.Error("Failed to activate project", "error", "project_id", projectId, "another project is already active")
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectId,
			"error":      "another project is already active",
		})
		return errors.New("another project is already active")
	}
	//	all shards must be inactive
	allShardsNotActive, err := a.checkAllShardsActive(projectId)
	if err != nil {
		logger.Logger.Error("Failed to check status of projects for projct activation", "project_id", projectId, "error", err)
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectId,
			"error":      "all project shards are not active",
		})
		return err
	}
	if allShardsNotActive == false {
		logger.Logger.Error("Failed to activate project", "error", "project_id", projectId, "All shards are not active")
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectId,
			"error":      "another project is already active",
		})
		return errors.New("All shards are not active")
	}
	err = a.ProjectRepo.ProjectActive(a.ctx, projectId)
	if err != nil {
		logger.Logger.Error("Failed to activate project", "project_id", projectId, "error", err)
		a.emitter.Error("Project activation failed", "application - Activateproject", map[string]string{
			"project_id": projectId,
			"error":      err.Error(),
		})
		return err
	}
	err = a.connManager.InitiateConnectionsAll(a.ctx)
	logger.Logger.Info("Successfully activated the project", "project_id", projectId)
	a.emitter.Info("Project activation successfull", "application - Activateproject", map[string]string{
		"project_id": projectId,
	})
	return nil

}

// project must be invalid , schema must be in draft state
func (a *App) UpdateProjectSchemaDraft(projectId string, schemaId string, ddlString string) error {
	ok, err := a.checkIfProjectInactive(projectId)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "project_id", projectId, "error", err)
		a.emitter.Error("Schema draft update failed", "application - UpdateProjectSchemaDraft", map[string]string{
			"project_id": projectId,
			"error":      "failed to check project status",
		})
		return err
	}

	if !ok {
		a.emitter.Error("Schema draft update failed", "application - UpdateProjectSchemaDraft", map[string]string{
			"project_id": projectId,
			"error":      "project is active",
		})
		return errors.New("project must be inactive to modify schema")
	}

	ok, err = a.checkIfSchemaDraft(schemaId)
	if err != nil {
		logger.Logger.Error("Failed to fetch schema state", "schema_id", schemaId, "error", err)
		a.emitter.Error("Schema draft update failed", "application - UpdateProjectSchemaDraft", map[string]string{
			"schema_id": schemaId,
			"error":     "failed to fetch schema state",
		})
		return err
	}

	if !ok {
		a.emitter.Error("Schema draft update failed", "application - UpdateProjectSchemaDraft", map[string]string{
			"schema_id": schemaId,
			"error":     "schema is not in draft state",
		})
		return errors.New("only draft schemas can be updated")
	}

	err = a.ProjectSchemaRepo.ProjectSchemaUpdateDraft(a.ctx, schemaId, ddlString)
	if err != nil {
		logger.Logger.Error("Failed to update schema draft", "schema_id", schemaId, "error", err)
		a.emitter.Error("Schema draft update failed", "application - UpdateProjectSchemaDraft", map[string]string{
			"schema_id": schemaId,
			"error":     err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully updated schema draft", "schema_id", schemaId)
	a.emitter.Info("Schema draft updated successfully", "application - UpdateProjectSchemaDraft", map[string]string{
		"schema_id": schemaId,
	})

	return nil
}

func (a *App) GetShards(projectID string) ([]repository.Shard, error) {
	result, err := a.ShardRepo.ShardList(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to get shards", "project_id", projectID, "error", err)
		a.emitter.Error("Failed to get shards", "get_shards", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}
	return result, nil
}

func (a *App) DeleteConnection(shardID string) error {
	err := a.ShardConnectionRepo.ConnectionRemove(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to remove shard connection details", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard connection deletion failed", "application - DeleteConnection", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
	}
	logger.Logger.Info("Successfully deleted shard connection details", "shard_id", shardID)
	a.emitter.Info("Shard connection deletion successfull", "application - DeleteConnection", map[string]string{
		"shard_id": shardID,
	})
	return nil
}

// CANNOT_DELETE_ACTIVE_SHARD
func (a *App) DeleteShard(shardID string) (string, error) {
	isInactive, err := a.checkIfShardInactive(shardID)
	if err != nil {
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return "", err
	}
	if isInactive {
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    "cannot delete active shard",
		})
		return "CANNOT_DELETE_ACTIVE_SHARD", nil
	}
	err = a.DeleteConnection(shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard connection", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    "failed to delete shard connection",
		})
		return "", err
	}

	err = a.ShardRepo.ShardDelete(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to delete shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard deletion failed", "application - DeleteShard", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return "", err
	}
	logger.Logger.Info("Successfully deleted shard", "shard_id", shardID)
	a.emitter.Info("Shard deletion successful", "application - DeleteShard", map[string]string{
		"shard_id": shardID,
	})
	return "DELETED", nil
}

func (a *App) FetchConnectionInfo(shardID string) (repository.ShardConnection, error) {
	conn, err := a.ShardConnectionRepo.GetConnectionByShardID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fecth sahrd connection infomation", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard connection info fetching failed", "application - FetchConnectionInfo", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return repository.ShardConnection{}, err
	}
	return conn, nil
}

// Retry existing shard connections
// Fetch the project ID associated with the shard
// Check shard health (connectivity, status, etc.)
// Activate the shard in the repository
func (a *App) ActivateShard(shardID string) error {
	////Retry existing shard connections
	err := a.RetryShardConnections(a.ctx)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"shard_id": shardID,
			"error":    "retry mechanism failed",
		})
	}
	// Fetch the project ID associated with the shard
	projectID, err := a.ShardRepo.FetchShardProjectID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"shard_id": shardID,
			"error":    "project fetching failed for health checking",
		})
		return err
	}
	//Check shard health (connectivity, status, etc.)
	isConnected, err := a.checkShardHealth(a.ctx, projectID, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "projectid", projectID, "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      "failed for checking shard health",
		})
		return err
	}
	if !isConnected {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", "shard connection not available")
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      "shard connection failed",
		})
		return err
	}
	//// Activate the shard in the repository
	err = a.ShardRepo.ShardActivate(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to activate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Shard Activation failed", "application - ActivateShard", map[string]string{
			"project_id": projectID,
			"shard_id":   shardID,
			"error":      err.Error(),
		})
		return err
	}

	logger.Logger.Info("Successfully activated shard", "shard_id", shardID)
	a.emitter.Info("Shard activation successfull", "application - ActivateShard", map[string]string{
		"shard_id": shardID,
	})

	return nil
}

// Deactivates a shard
// Finds its project
// Fetches all shards of that project
// If any shard is inactive → deactivates the whole project
// Emits a system event
func (a *App) DeactivateShard(shardID string) error {
	if err := a.ShardRepo.ShardDeactivate(a.ctx, shardID); err != nil {
		logger.Logger.Error("Failed to deactivate shard", "shard_id", shardID, "error", err)
		a.emitter.Error("Failed to deactivate shard", "deactivate_shard", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return err
	}
	projectId, err := a.ShardRepo.FetchShardProjectID(a.ctx, shardID)
	if err != nil {
		logger.Logger.Error("Failed to fetch shard project ID", "shard_id", shardID, "error", err)
		a.emitter.Error("Failed to fetch shard project ID", "fetch_shard_project_id", map[string]string{
			"shard_id": shardID,
			"error":    err.Error(),
		})
		return err
	}
	shards, err := a.GetShards(projectId)
	if err != nil {
		logger.Logger.Error("Failed to get shards", "project_id", projectId, "error", err)
		a.emitter.Error("Failed to get shards", "get_shards", map[string]string{
			"project_id": projectId,
			"shard_id":   shardID,
			"error":      err.Error(),
		})
		return err
	}
	for _, shd := range shards {
		if shd.Status == "inactive" {
			if err := a.ProjectRepo.ProjectDeactivate(a.ctx, projectId); err != nil {
				logger.Logger.Error("Failed to deactivate project", "project_id", projectId, "shard_id", shardID, "error", err)
				a.emitter.Error("Failed to deactivate project", "deactivate_project", map[string]string{
					"project_id": projectId,
					"shard_id":   shardID,
					"error":      err.Error(),
				})
				return err
			}
			break
		}
	}
	logger.Logger.Info("Automatic project deactivation triggered", "project_id", projectId)
	a.emitter.Info("Automatic project deactivation triggered", "automatic_project_deactivation", map[string]string{
		"project_id": projectId,
	})
	logger.Logger.Info("Shard deactivated successfully", "shard_id", shardID)
	a.emitter.Info("Shard deactivated successfully", "deactivate_shard", map[string]string{
		"shard_id":   shardID,
		"project_id": projectId,
	})
	return nil
}

func (a *App) AddConnection(connectionInfo *repository.ShardConnection) error {
	err := a.ShardConnectionRepo.ConnectionCreate(a.ctx, connectionInfo)
	if err != nil {
		logger.Logger.Error("Failed to add shard connection details", "shard_id", connectionInfo.ShardId, "error", err)
		a.emitter.Error("shard connection addition failed", "application - AddConnection", map[string]string{
			"shard_id": connectionInfo.ShardId,
			"error":    err.Error(),
		})
		return err
	}
	logger.Logger.Info("Successfully added shard connection details", "shard_id", connectionInfo.ShardId)
	a.emitter.Info("shard connection addition successfull", "application - AddConnection", map[string]string{
		"shard_id": connectionInfo.ShardId,
	})

	return nil
}

func (a *App) UpdateConnection(connInfo repository.ShardConnection) error {
	err := a.ShardConnectionRepo.ConnectionUpdate(a.ctx, connInfo)
	if err != nil {
		logger.Logger.Error("Failed to update shard connection details", "shard_id", connInfo.ShardId, "error", err)
		a.emitter.Error("Shard connection updation failed", "application - UpdateConnection", map[string]string{
			"shard_id": connInfo.ShardId,
			"error":    err.Error(),
		})
		return err
	}
	logger.Logger.Info("Successfully updated shard connection details", "shard_id", connInfo.ShardId)
	a.emitter.Info("Shard connection updation successfull", "application - UpdateConnection", map[string]string{
		"shard_id": connInfo.ShardId,
	})
	return nil
}

func (a *App) DeleteAllShards(projectID string) error {
	err := a.ShardRepo.ShardDeleteAll(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to delete all shards", "project_id", projectID, "error", err)
		a.emitter.Error("Failed to delete all shards", "delete_all_shards", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}
	logger.Logger.Info("All shards deleted successfully", "project_id", projectID)
	a.emitter.Info("All shards deleted successfully", "delete_all_shards", map[string]string{
		"project_id": projectID,
	})
	return nil
}

func (a *App) GetFailedShardExecutions(schemaID string) ([]repository.SchemaExecutionStatus, error) {
	statuAll, err := a.SchemaExecutionStatusRepo.ExecutionRecordsFetchStatusFailed(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch execution status of all failed records of schema", "schema_id", schemaID, "error", err)
		a.emitter.Error("Failed shard execution status fetching failed", "application - GetFailedShardExecutions", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return nil, err
	}
	return statuAll, err
}

func (a *App) GetShardStatus(shardId string) (string, error) {
	status, err := a.ShardRepo.FetchShardStatus(a.ctx, shardId)
	if err != nil {
		logger.Logger.Error("Failed to get shard status", "shard_id", shardId, "error", err)
		a.emitter.Error("Failed to get shard status", "get_shard_status", map[string]string{
			"shard_id": shardId,
			"error":    err.Error(),
		})
		return "", err
	}
	return status, nil
}

func (a *App) Deactivateproject(projectID string) error {
	err := a.ProjectRepo.ProjectDeactivate(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to deactivate project", "project_id", projectID, "error", err)
		a.emitter.Error("Project deletion failed", "application - Deactivateproject", map[string]string{
			"project_Id": projectID,
			"error":      err.Error(),
		})
	}

	logger.Logger.Info("Successfully deactivated the project", "project_id", projectID)
	a.emitter.Info("Project deletion successfull", "application - Deactivateproject", map[string]string{
		"project_Id": projectID,
	})
	return nil
}

func (a *App) FetchProjectStatus(projectID string) (string, error) {
	status, err := a.ProjectRepo.FetchProjectStatus(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "project_id", projectID, "error", err)
		a.emitter.Error("Project status fetching failed", "application - FetchProjectStatus", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return "", err
	}
	return status, nil

}

func (a *App) CreateSchemaDraft(projectID string, ddlSQL string) (*repository.ProjectSchema, error) {
	schema, err := a.ProjectSchemaRepo.CreateProjectSchemaDraft(a.ctx, projectID, ddlSQL)
	if err != nil {
		logger.Logger.Error("Failed to create schema draft", "project_id", projectID, "error", err)
		a.emitter.Error("Schema draft creation failed", "application - CreateSchemaDraft", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}
	logger.Logger.Info("Succesfully created schema draft of project", "project_id", projectID)
	a.emitter.Info("Schema draft creation successfull", "application - CreateSchemaDraft", map[string]string{
		"project_id": projectID,
	})
	return schema, nil
}

func (a *App) GetCurrentSchema(projectID string) (*repository.ProjectSchema, error) {
	schema, err := a.ProjectSchemaRepo.ProjectSchemaGetLatest(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch latest schema of project", "project_id", projectID, "error", err)
		a.emitter.Error("Current schema fetching failed", "application - GetCurrentSchema", map[string]string{
			"projecy_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}
	return schema, nil

}

func (a *App) GetSchemaHistory(projectID string) ([]repository.ProjectSchema, error) {
	history, err := a.ProjectSchemaRepo.ProjectSchemaFetchHistory(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project schema history", "project_id", projectID, "error", err)
		a.emitter.Error("Schema history fetching failed", "application - GetSchemaHistory", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}
	return history, nil
}

func (a *App) DeleteSchemaDraft(schemaID string) error {
	err := a.ProjectSchemaRepo.ProjectSchemaDeleteDraft(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to delete project schema draft", "schema_id", schemaID, "error", err)
		a.emitter.Error("Schema deletion failed", "application - DeleteSchemaDraft", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return err
	}
	logger.Logger.Info("Successfully deleted project schema draft", "schema_id", schemaID)
	a.emitter.Info("Schema deletion successfull", "application - DeleteSchemaDraft", map[string]string{
		"schema_id": schemaID,
	})
	return nil
}

func (a *App) GetSchemaExecutionStatus(schemaID string) ([]repository.SchemaExecutionStatus, error) {
	statuAll, err := a.SchemaExecutionStatusRepo.ExecutionRecordsFetchStatusAll(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch execution status of all records of schema", "schema_id", schemaID, "error", err)
		a.emitter.Error("Schema execution status fetching failed", "application - GetSchemaExecutionStatus", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return nil, err
	}
	return statuAll, err
}

func (a *App) GetProjectSchemaStatus(schemaID string) (string, error) {
	status, err := a.ProjectSchemaRepo.ProjectSchemaGetState(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Fialed to fetch status of a schema", "schema_id", schemaID, "error", err)
		a.emitter.Error("Project schema status fetching failed", "application - GetProjectSchemaStatus", map[string]string{
			"schema_id": schemaID,
			"error":     err.Error(),
		})
		return "", err
	}

	return status, nil
}

func (a *App) FetchShardKeys(projectID string) ([]repository.ShardKeys, error) {
	keys, err := a.ShardKeysRepo.FetchShardKeysByProjectID(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("failed to fetch shard keys", "project_id", projectID, "error", err)
		a.emitter.Error("Shard key fetching failed", "application - FetchShardKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return nil, err
	}
	return keys, nil
}

func (a *App) ReplaceShardKeys(projectID string, keys []repository.ShardKeyRecord) error {
	err := a.ShardKeysRepo.ReplaceShardKeysForProject(a.ctx, projectID, keys)
	if err != nil {
		logger.Logger.Error("filed to replace shard keys", "projectID", projectID, "error", err)
		a.emitter.Error("Shard key replacing failed", "application - ReplaceShardKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}
	logger.Logger.Info("successfully replaced shard keys", "projectID", projectID)
	a.emitter.Info("Shard key replacing successfull", "application - ReplaceShardKeys", map[string]string{
		"project_id": projectID,
	})
	return nil
}

func (a *App) MonitorShards(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Shard monitor stopped")
			a.emitter.Info("Shard monitor stopper", "application - MonitorShards", map[string]string{})
			return

		case <-ticker.C:
			a.checkAllShards(ctx)
		}
	}
}

func (a *App) RetryShardConnections(ctx context.Context) error {
	err := a.connManager.InitiateConnectionsAll(a.ctx)
	if err != nil {
		logger.Logger.Error("shard connection retry mechanism failed", "error", err)
		a.emitter.Error("Shard connection retry mechanism failed", "application - RetryShardConnections", map[string]string{
			"error": err.Error(),
		})
		return err
	}
	a.emitter.Info("Shard connection retry mechanism successfull", "application - RetryShardConnections", map[string]string{})
	return nil
}

func (a *App) execDDLonShard(projectID string, shardID string, ddl string) error {
	db, err := a.connStore.Get(projectID, shardID)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(a.ctx, ddl)
	return err
}

func (a *App) ApplyDDL(projectID, ddl string) error {
	if err := a.schemaService.ApplyDDL(a.ctx, projectID, ddl); err != nil {
		logger.Logger.Error("Failed to apply DDL", "project_id", projectID, "error", err)
		a.emitter.Error("Failed to apply DDL", "apply_ddl", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}
	logger.Logger.Info("DDL applied", "project_id", projectID)
	a.emitter.Info("DDL applied", "apply_ddl", map[string]string{
		"project_id": projectID,
	})
	return nil
}

// project must be invctive , schema must be in draft state -> (remove destructive query ) ->
// recompute shard keys -> commit draft
func (a *App) CommitSchemaDraft(projectID string, schemaID string) error {
	ok, err := a.checkIfProjectInactive(projectID)
	if err != nil {
		logger.Logger.Error("Failed to fetch project status", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to check project status",
		})
		return err
	}

	if !ok {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "project is active",
		})
		return errors.New("project must be inactive to modify schema")
	}

	ok, err = a.checkIfSchemaDraft(schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch schema state", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to fetch schema state",
		})
		return err
	}
	if !ok {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "schema should be in draft before commit",
		})
		return errors.New("schema must be in draft state to commit")
	}

	inFlight, err := a.checkIfSchemaInFlight(projectID)
	if err != nil {
		logger.Logger.Error("Failed to check schema in-flight status", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to check schema in-flight status",
		})
		return err
	}
	if inFlight {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "another schema change is already in progress",
		})
		return errors.New("another schema change is already in progress")
	}

	projectSchema, err := a.ProjectSchemaRepo.ProjectSchemaGetBySchemaID(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to fetch schema by id", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to fetch schema infomation",
		})
		return err
	}

	destructive, err := a.checkIfDDLDestructive(projectID, projectSchema.DDL_SQL)
	if err != nil {
		logger.Logger.Error("Failed to validate destructive DDL", "project_id", projectID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "failed to validate destructive DDL",
		})
		return err
	}
	if destructive {
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "destructive ddl in query",
		})
		return errors.New("destructive DDL is not allowed after initial schema")
	}

	logger.Logger.Info("Applying committed schema to metadata", "project_id", projectID, "schema_id", schemaID)
	a.emitter.Info("applying committed schema to metadata", "application - CommitSchemaDraft", map[string]string{
		"project_id": projectID,
	})
	err = a.schemaService.ApplyDDL(a.ctx, projectID, projectSchema.DDL_SQL)
	if err != nil {
		logger.Logger.Error("Failed to apply schema to metadata", "project_id", projectID, "schema_id", schemaID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "Failed to apply schema to metadata" + err.Error(),
		})
		return err
	}
	err = a.ProjectSchemaRepo.ProjectSchemaCommitDraft(a.ctx, schemaID)
	if err != nil {
		logger.Logger.Error("Failed to commit project schema", "project_id", projectID, "schema_id", schemaID, "error", err)
		a.emitter.Error("Draft schema commit failed", "application - CommitSchemaDraft", map[string]string{
			"project_id": projectID,
			"error:":     "Failed to commit project schema" + err.Error(),
		})
		return err
	}
	logger.Logger.Info("Successfully committed project schema", "project_id", projectID, "schema_id", schemaID)
	a.emitter.Info("Schema draft commit successfull", "application - CommitSchemaDraft", map[string]string{
		"project_id": projectID,
	})
	return nil
}

// project must be inactive, shard key recomputation
func (a *App) RecomputeKeys(projectID string) error {
	isProjectActive, err := a.checkIfProjectInactive(projectID)
	if err != nil {
		logger.Logger.Error("failed to recompute shard keys", "project_id", projectID, "error", err)
		a.emitter.Error("Shard key recomputation failed", "application -RecomputeKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}
	if !isProjectActive {
		a.emitter.Error("Shard key recomputation failed", "application -RecomputeKeys", map[string]string{
			"project_id": projectID,
			"error":      "project must be inactive before shard key recomputation",
		})
		return err
	}
	err = a.InferenceService.RunForProject(a.ctx, projectID)
	if err != nil {
		logger.Logger.Error("shard key inference failed", "project_id", projectID, "error", err)
		a.emitter.Error("Shard key recomputation failed", "application -RecomputeKeys", map[string]string{
			"project_id": projectID,
			"error":      err.Error(),
		})
		return err
	}
	logger.Logger.Info("shard key inference completed successfully", "project_id", projectID)
	a.emitter.Info("Shard key recomputation successfull	", "application -RecomputeKeys", map[string]string{
		"project_id": projectID,
	})

	return nil

}

func (a *App) ExecuteSQL(projectID string, sqlText string) ([]executor.ExecutionResult, error) {
	plan, err := a.shardRouter.RouteSQL(
		a.ctx,
		projectID,
		sqlText,
	)
	if err != nil {
		logger.Logger.Error("Filed to route and execute query", "project_id", projectID, "error", err)
		a.emitter.Error("Query execution failed", "application - ExecuteSQL", map[string]string{
			"project_id": projectID,
			"error":      "query router:" + err.Error(),
		})
		return nil, err
	}
	result, err := a.ExecutorService.Execute(
		a.ctx,
		projectID,
		sqlText,
		plan,
	)
	if err != nil {
		logger.Logger.Error("failed to execute query", "project_id", projectID, "error", err)
		logger.Logger.Error("Filed to route and execute query", "project_id", projectID, "error", err)
		a.emitter.Error("Query execution failed", "application - ExecuteSQL", map[string]string{
			"project_id": projectID,
			"error":      "query executor:" + err.Error(),
		})
		return nil, err
	}
	return result, nil
}

//ExecuteProjectSchema
//RetrySchemaExecution
//GetSchemaCapabilities
