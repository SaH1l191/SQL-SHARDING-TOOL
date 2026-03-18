package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sqlsharder/internal/handler"
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

func (r *Router) RegisterProjectRoutes(h *handler.ProjectHandler) {
	v1 := r.engine.Group("/api/v1")

	projects := v1.Group("/projects")
	{
		projects.POST("/create", h.CreateProject)
		projects.GET("/list", h.GetProjects)
		projects.DELETE("/:id", h.DeleteProject)
		projects.PUT("/:id/activate", h.ActivateProject)
		projects.PUT("/:id/deactivate", h.DeactivateProject)
		projects.GET("/:id/status", h.GetProjectStatus)
	}
}
func (r *Router) RegisterShardRoutes(h *handler.ShardHandler) {
	v1 := r.engine.Group("/api/v1")

	shards := v1.Group("/shards")
	{
		shards.POST("/create", h.CreateShard)
		shards.GET("/list", h.GetShards)
		shards.DELETE("/:id", h.DeleteShard)
		shards.PUT("/:id/activate", h.ActivateShard)
		shards.PUT("/:id/deactivate", h.DeactivateShard)
		shards.GET("/:id/status", h.GetShardStatus)
	}
}
