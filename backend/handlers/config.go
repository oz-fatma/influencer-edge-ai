package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/influencer-edge-ai/backend/config"
)

type ConfigHandler struct {
	cfg *config.Config
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"app_name":    h.cfg.AppName,
		"app_version": h.cfg.AppVersion,
		"environment": h.cfg.AppEnv,
		"features": gin.H{
			"auth":               true,
			"web_llm_scores":     true,
			"influencer_analysis": true,
		},
		"supported_platforms": []string{
			"instagram", "tiktok", "youtube", "twitter", "linkedin", "other",
		},
	})
}

func (h *ConfigHandler) GetHealthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":           "ok",
		"database":         "postgres",
		"cors_origins_set": len(h.cfg.CORSAllowedOrigins) > 0,
	})
}
