package router

import (
	"net/http"
	"sqlsharder/internal/handler"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func New() *Router {
	return &Router{engine: gin.Default()}
}

func (r *Router) Engine() http.Handler {
	return r.engine
}

func (r *Router) RegisterHealthRoute() {
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// RegisterProjectRoutes wires all project endpoints.
// FIX: removed /create and /list path suffixes — REST convention is
//
//	POST /projects (not POST /projects/create) and GET /projects (not /projects/list).
func (r *Router) RegisterProjectRoutes(h *handler.ProjectHandler) {
	v1 := r.engine.Group("/api/v1")

	projects := v1.Group("/projects")
	{
		projects.POST("", h.CreateProject)
		projects.GET("", h.GetProjects)
		projects.DELETE("/:id", h.DeleteProject)
		projects.PUT("/:id/activate", h.ActivateProject)
		projects.PUT("/:id/deactivate", h.DeactivateProject)
		projects.GET("/:id/status", h.GetProjectStatus)
	}
}

// RegisterShardRoutes wires all shard endpoints.
func (r *Router) RegisterShardRoutes(h *handler.ShardHandler) {
	v1 := r.engine.Group("/api/v1")

	shards := v1.Group("/shards")
	{
		shards.POST("", h.CreateShard)
		shards.GET("", h.GetShards) //query : project_id
		shards.DELETE("/:id", h.DeleteShard)
		shards.PUT("/:id/activate", h.ActivateShard)
		shards.PUT("/:id/deactivate", h.DeactivateShard)
		shards.GET("/:id/status", h.GetShardStatus)
	}
}

func (r *Router) RegisterProjectSchemaRoutes(h *handler.ProjectSchemaHandler) {
	v1 := r.engine.Group("/api/v1")

	schemas := v1.Group("/schemas")
	{
		schemas.POST("", h.CreateSchema)
		schemas.GET("", h.GetSchemasByProject)
		schemas.GET("/:id", h.GetSchemaByID)
		schemas.PUT("/:id", h.UpdateSchema)
		schemas.DELETE("/:id", h.DeleteSchema)
		schemas.PUT("/:id/commit", h.CommitSchema)
		schemas.GET("/:id/status", h.GetSchemaStatus)
		schemas.PUT("/:id/applying", h.SetApplying)
		schemas.PUT("/:id/state", h.UpdateSchemaState)
	}

	projectSchemas := v1.Group("/projects/schemas")
	{
		projectSchemas.GET("/latest", h.GetLatestSchema)
		projectSchemas.GET("/applied", h.GetAppliedSchema)
		projectSchemas.GET("/versions", h.GetProjectSchemaVersions)
		projectSchemas.GET("/pending", h.GetPendingSchema)
	}
}

// RegisterShardConnectionRoutes wires all shard connection endpoints.
func (r *Router) RegisterShardConnectionRoutes(h *handler.ShardConnectionHandler) {
	v1 := r.engine.Group("/api/v1")

	shardConnections := v1.Group("/shard-connections")
	{
		shardConnections.POST("", h.CreateShardConnection)
		shardConnections.GET("/:id", h.GetShardConnection)
		shardConnections.PUT("/:id", h.UpdateShardConnection)
		shardConnections.DELETE("/:id", h.DeleteShardConnection)
	}
}
