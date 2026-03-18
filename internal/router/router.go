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
 