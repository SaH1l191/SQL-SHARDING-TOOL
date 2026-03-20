package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"sqlsharder/internal/service"
	"sqlsharder/pkg/logger"
)

type ShardConnectionHandler struct {
	service *service.ShardConnectionService
}

func NewShardConnectionHandler(svc *service.ShardConnectionService) *ShardConnectionHandler {
	return &ShardConnectionHandler{service: svc}
}

func (h *ShardConnectionHandler) CreateShardConnection(c *gin.Context) {
	var req service.CreateShardConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("Invalid input", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	conn, err := h.service.CreateShardConnection(c.Request.Context(), &req)
	if err != nil {
		logger.Logger.Error("Failed to create shard connection", "error", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, conn)
}

func (h *ShardConnectionHandler) GetShardConnection(c *gin.Context) {
	shardID := c.Param("id")

	conn, err := h.service.GetShardConnection(c.Request.Context(), shardID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, conn)
}

func (h *ShardConnectionHandler) UpdateShardConnection(c *gin.Context) {
	shardID := c.Param("id")

	var req service.UpdateShardConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("Invalid input", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	conn, err := h.service.UpdateShardConnection(c.Request.Context(), shardID, &req)
	if err != nil {
		logger.Logger.Error("Failed to update shard connection", "error", err)
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, conn)
}

func (h *ShardConnectionHandler) DeleteShardConnection(c *gin.Context) {
	shardID := c.Param("id")

	err := h.service.DeleteShardConnection(c.Request.Context(), shardID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handleError maps service sentinel errors to HTTP status codes.
func (h *ShardConnectionHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, service.ErrShardConnectionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "shard connection not found"})

	default:
		logger.Logger.Error("Unhandled error", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
