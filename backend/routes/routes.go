package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/influencer-edge-ai/backend/config"
	"github.com/influencer-edge-ai/backend/handlers"
	"github.com/influencer-edge-ai/backend/middleware"
	"github.com/influencer-edge-ai/backend/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, db *gorm.DB, redis *redis.Client, cfg *config.Config) {
	common := handlers.NewCommonHandler(cfg, db, redis)
	configH := handlers.NewConfigHandler(cfg)
	auth := handlers.NewAuthHandler(db, cfg)
	score := handlers.NewScoreHandler(db)
	monitoring := handlers.NewMonitoringHandler(store.NewLLMMetricsStore(redis))

	jwtAuth := middleware.JWTAuth(cfg)

	// Common [2]
	r.GET("/health", common.Health)
	r.GET("/version", common.Version)

	// Config [2]
	r.GET("/config", configH.GetConfig)
	r.GET("/health-config", configH.GetHealthConfig)

	// Auth [8] — public
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", auth.Register)
		authGroup.POST("/login", auth.Login)
		authGroup.POST("/logout", auth.Logout)
		authGroup.POST("/refresh-token", auth.RefreshToken)
	}

	// Auth [8] — protected
	authProtected := r.Group("/auth")
	authProtected.Use(jwtAuth)
	{
		authProtected.GET("/profile", auth.GetProfile)
		authProtected.PUT("/profile", auth.UpdateProfile)
		authProtected.PUT("/change-password", auth.ChangePassword)
		authProtected.DELETE("/account", auth.DeleteAccount)
	}

	// WebMLC-LLM [8] — protected
	api := r.Group("/api")
	api.Use(jwtAuth)
	{
		api.POST("/scores", score.CreateScore)
		api.GET("/scores", score.GetScores)
		api.GET("/scores/:id", score.GetScoreByID)
		api.PUT("/scores/:id", score.UpdateScore)
		api.DELETE("/scores/:id", score.DeleteScore)

		api.POST("/analyses", score.CreateAnalysis)
		api.GET("/analyses", score.ListAnalyses)
		api.GET("/influencer-analysis/:id", score.GetInfluencerAnalysis)

		api.POST("/llm-metrics", monitoring.RecordLLMMetric)
		api.GET("/monitoring/stats", monitoring.GetMonitoringStats)
	}
}
