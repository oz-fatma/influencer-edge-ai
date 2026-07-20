package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/influencer-edge-ai/backend/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CommonHandler struct {
	cfg   *config.Config
	db    *gorm.DB
	redis *redis.Client
}

func NewCommonHandler(cfg *config.Config, db *gorm.DB, redis *redis.Client) *CommonHandler {
	return &CommonHandler{cfg: cfg, db: db, redis: redis}
}

func (h *CommonHandler) Health(c *gin.Context) {
	checks := gin.H{}
	healthy := true

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		checks["database"] = "down"
		healthy = false
	} else {
		checks["database"] = "up"
	}

	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		checks["redis"] = "down"
		healthy = false
	} else {
		checks["redis"] = "up"
	}

	status := "healthy"
	code := http.StatusOK
	if !healthy {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"status": status,
		"checks": checks,
	})
}

func (h *CommonHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":    h.cfg.AppName,
		"version": h.cfg.AppVersion,
		"env":     h.cfg.AppEnv,
	})
}
