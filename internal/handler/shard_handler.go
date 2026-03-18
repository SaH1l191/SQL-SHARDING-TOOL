package handler

import (
	"errors"
	"net/http"
	"sqlsharder/internal/service"
	"sqlsharder/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ShardHandler struct {
	service *service.ShardService
}

func NewShardHandler(svc *service.ShardService) *ShardHandler {
	return &ShardHandler{service: svc}
}

func (h *ShardHandler) CreateShard(c *gin.Context) {
	var req service.CreateShardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("Invalid input", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	shard, err := h.service.CreateShard(c.Request.Context(),&req)
	if err != nil {
		logger.Logger.Error("Failed to create shard", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shard: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shard)
}

func (h *ShardHandler) GetShards(c *gin.Context) {
	projectID := c.Query("project_id")

	shards, err := h.service.GetShards(c.Request.Context(), projectID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, shards)
}

func (h *ShardHandler) DeleteShard(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteShard(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ShardHandler) ActivateShard(c *gin.Context) {
	id := c.Param("id")

	err := h.service.ActivateShard(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "shard activated"})
}

func (h *ShardHandler) DeactivateShard(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeactivateShard(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "shard deactivated"})
}

func (h *ShardHandler) GetShardStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.service.GetShardStatus(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"shard_id": id,
		"status":   status,
	})
}

// centralized error handler
func (h *ShardHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrShardNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "shard not found"})

	case errors.Is(err, service.ErrShardDeleteBlocked):
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete active shard"})

	default:
		logger.Logger.Error("Unhandled error", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
