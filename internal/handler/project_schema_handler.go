package handler

import (
	"errors"
	"net/http"
	"sqlsharder/internal/service"

	"github.com/gin-gonic/gin"
)

type ProjectSchemaHandler struct {
	service *service.ProjectSchemaService
}

func NewProjectSchemaHandler(svc *service.ProjectSchemaService) *ProjectSchemaHandler {
	return &ProjectSchemaHandler{service: svc}
}

// CreateSchema POST /api/v1/schemas
func (h *ProjectSchemaHandler) CreateSchema(c *gin.Context) {
	var req service.CreateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schema, err := h.service.CreateSchema(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, schema)
}

// GetSchemasByProject GET /api/v1/schemas?project_id=xxx
func (h *ProjectSchemaHandler) GetSchemasByProject(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id query param is required"})
		return
	}

	schemas, err := h.service.GetSchemasByProject(c.Request.Context(), projectID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, schemas)
}

// GetSchemaByID GET /api/v1/schemas/:id
func (h *ProjectSchemaHandler) GetSchemaByID(c *gin.Context) {
	id := c.Param("id")

	schema, err := h.service.GetSchemaByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, schema)
}

// UpdateSchema PUT /api/v1/schemas/:id
func (h *ProjectSchemaHandler) UpdateSchema(c *gin.Context) {
	var req service.UpdateSchemaRequest
	// Bind schema_id from the URL param, not from the body
	req.SchemaID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schema, err := h.service.UpdateSchema(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, schema)
}

// DeleteSchema DELETE /api/v1/schemas/:id
func (h *ProjectSchemaHandler) DeleteSchema(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteSchema(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// CommitSchema PUT /api/v1/schemas/:id/commit  (draft → pending)
// FIX: renamed from ActivateSchema — "activate" implies running DDL.
//      "commit" correctly means "lock in this draft and queue it for execution".
func (h *ProjectSchemaHandler) CommitSchema(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.CommitSchema(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schema committed, pending execution"})
}

// GetSchemaStatus GET /api/v1/schemas/:id/status
func (h *ProjectSchemaHandler) GetSchemaStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.service.GetSchemaStatus(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"schema_id": id, "status": status})
}

// GetLatestSchema GET /api/v1/projects/:project_id/schemas/latest
// FIX: uses path param :project_id, not c.Param("project_id") on a flat route.
//      This route is registered under /projects/:project_id/schemas/latest.
func (h *ProjectSchemaHandler) GetLatestSchema(c *gin.Context) {
	projectID := c.Param("project_id")

	schema, err := h.service.GetLatestSchema(c.Request.Context(), projectID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, schema)
}

// GetAppliedSchema GET /api/v1/projects/:project_id/schemas/applied
func (h *ProjectSchemaHandler) GetAppliedSchema(c *gin.Context) {
	projectID := c.Param("project_id")

	schema, err := h.service.GetAppliedSchema(c.Request.Context(), projectID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, schema)
}

// handleError maps service sentinel errors to HTTP status codes.
func (h *ProjectSchemaHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSchemaNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
	case errors.Is(err, service.ErrInvalidState):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}