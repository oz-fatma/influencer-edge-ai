package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/influencer-edge-ai/backend/middleware"
	"github.com/influencer-edge-ai/backend/store"
	"github.com/influencer-edge-ai/backend/utils"
)

type MonitoringHandler struct {
	metrics *store.LLMMetricsStore
}

func NewMonitoringHandler(metrics *store.LLMMetricsStore) *MonitoringHandler {
	return &MonitoringHandler{metrics: metrics}
}

type recordLLMMetricRequest struct {
	InfluencerName string `json:"influencer_name" binding:"required"`
	LatencyMs      int64  `json:"latency_ms" binding:"required"`
	Status         string `json:"status" binding:"required"`
	Model          string `json:"model" binding:"required"`
}

func (h *MonitoringHandler) RecordLLMMetric(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req recordLLMMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := utils.ValidateInfluencerName(req.InfluencerName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status != "success" && status != "error" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be success or error"})
		return
	}

	if req.LatencyMs < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latency_ms must be non-negative"})
		return
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}
	if len(model) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model must be at most 128 characters"})
		return
	}

	metric := store.LLMCallMetric{
		InfluencerName: strings.TrimSpace(req.InfluencerName),
		LatencyMs:      req.LatencyMs,
		Status:         status,
		Model:          model,
	}

	if err := h.metrics.Record(c.Request.Context(), userID, &metric); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record metric"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "metric recorded",
		"metric":  metric,
	})
}

func (h *MonitoringHandler) GetMonitoringStats(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.metrics.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch monitoring stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
