package handler

import (
	"errors"
	"net/http"
	"sqlsharder/internal/service"
	"sqlsharder/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: svc}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req service.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("Invalid input", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	project, err := h.service.CreateProject(c.Request.Context(), &req)
	if err != nil {
		logger.Logger.Error("Failed to create project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) GetProjects(c *gin.Context) {
	projects, err := h.service.GetProjects(c.Request.Context())
	if err != nil {
		logger.Logger.Error("Failed to retrieve projects", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteProject(c.Request.Context(), id); err != nil {
		logger.Logger.Error("Failed to delete project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) ActivateProject(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ActivateProject(c.Request.Context(), id); err != nil {
		logger.Logger.Error("Failed to activate project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project activated"})
}

func (h *ProjectHandler) DeactivateProject(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeactivateProject(c.Request.Context(), id); err != nil {
		logger.Logger.Error("Failed to deactivate project", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deactivated"})
}

func (h *ProjectHandler) GetProjectStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.service.GetProjectStatus(c.Request.Context(), id)
	if err != nil {
		logger.Logger.Error("Failed to get project status", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project_id": id, "status": status})
}

// handleError maps service sentinel errors to HTTP status codes.
// All handlers call this — error translation lives in exactly one place.
func (h *ProjectHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
