package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/influencer-edge-ai/backend/middleware"
	"github.com/influencer-edge-ai/backend/models"
	"github.com/influencer-edge-ai/backend/utils"
	"gorm.io/gorm"
)

type ScoreHandler struct {
	db *gorm.DB
}

func NewScoreHandler(db *gorm.DB) *ScoreHandler {
	return &ScoreHandler{db: db}
}

type createScoreRequest struct {
	InfluencerName  string  `json:"influencer_name" binding:"required"`
	Platform        string  `json:"platform" binding:"required"`
	OverallScore    float64 `json:"overall_score" binding:"required"`
	EngagementScore float64 `json:"engagement_score"`
	AudienceScore   float64 `json:"audience_score"`
	BrandFitScore   float64 `json:"brand_fit_score"`
	RawPayload      string  `json:"raw_payload"`
	Notes           string  `json:"notes"`
}

type updateScoreRequest struct {
	InfluencerName  *string  `json:"influencer_name"`
	Platform        *string  `json:"platform"`
	OverallScore    *float64 `json:"overall_score"`
	EngagementScore *float64 `json:"engagement_score"`
	AudienceScore   *float64 `json:"audience_score"`
	BrandFitScore   *float64 `json:"brand_fit_score"`
	RawPayload      *string  `json:"raw_payload"`
	Notes           *string  `json:"notes"`
}

type createAnalysisRequest struct {
	InfluencerName string `json:"influencer_name" binding:"required"`
	Platform       string `json:"platform" binding:"required"`
	AnalysisType   string `json:"analysis_type" binding:"required"`
	Summary        string `json:"summary" binding:"required"`
	Insights       string `json:"insights"`
	RawLLMOutput   string `json:"raw_llm_output"`
	ScoreID        *uint  `json:"score_id"`
}

func (h *ScoreHandler) CreateScore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := validateScoreInput(req.InfluencerName, req.Platform, req.OverallScore,
		req.EngagementScore, req.AudienceScore, req.BrandFitScore); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score := models.InfluencerScore{
		UserID:          userID,
		InfluencerName:  strings.TrimSpace(req.InfluencerName),
		Platform:        strings.ToLower(strings.TrimSpace(req.Platform)),
		OverallScore:    req.OverallScore,
		EngagementScore: req.EngagementScore,
		AudienceScore:   req.AudienceScore,
		BrandFitScore:   req.BrandFitScore,
		RawPayload:      req.RawPayload,
		Notes:           req.Notes,
	}

	if err := h.db.Create(&score).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create score"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"score": score})
}

func (h *ScoreHandler) GetScores(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	platform := strings.TrimSpace(c.Query("platform"))
	query := h.db.Where("user_id = ?", userID)
	if platform != "" {
		query = query.Where("platform = ?", strings.ToLower(platform))
	}

	var scores []models.InfluencerScore
	if err := query.Order("created_at DESC").Find(&scores).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch scores"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores, "count": len(scores)})
}

func (h *ScoreHandler) GetScoreByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score id"})
		return
	}

	var score models.InfluencerScore
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&score).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "score not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"score": score})
}

func (h *ScoreHandler) UpdateScore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score id"})
		return
	}

	var req updateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var score models.InfluencerScore
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&score).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "score not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch score"})
		return
	}

	if req.InfluencerName != nil {
		if err := utils.ValidateInfluencerName(*req.InfluencerName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		score.InfluencerName = strings.TrimSpace(*req.InfluencerName)
	}
	if req.Platform != nil {
		if err := utils.ValidatePlatform(*req.Platform); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		score.Platform = strings.ToLower(strings.TrimSpace(*req.Platform))
	}
	if req.OverallScore != nil {
		if err := utils.ValidateScore(*req.OverallScore); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		score.OverallScore = *req.OverallScore
	}
	if req.EngagementScore != nil {
		if err := utils.ValidateScore(*req.EngagementScore); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		score.EngagementScore = *req.EngagementScore
	}
	if req.AudienceScore != nil {
		if err := utils.ValidateScore(*req.AudienceScore); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		score.AudienceScore = *req.AudienceScore
	}
	if req.BrandFitScore != nil {
		if err := utils.ValidateScore(*req.BrandFitScore); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		score.BrandFitScore = *req.BrandFitScore
	}
	if req.RawPayload != nil {
		score.RawPayload = *req.RawPayload
	}
	if req.Notes != nil {
		score.Notes = *req.Notes
	}

	if err := h.db.Save(&score).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update score"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"score": score})
}

func (h *ScoreHandler) DeleteScore(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid score id"})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.InfluencerScore{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete score"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "score not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "score deleted successfully"})
}

func (h *ScoreHandler) GetInfluencerAnalysis(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis id"})
		return
	}

	var analysis models.InfluencerAnalysis
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&analysis).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analysis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

func (h *ScoreHandler) ListAnalyses(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	platform := strings.TrimSpace(c.Query("platform"))
	query := h.db.Where("user_id = ?", userID)
	if platform != "" {
		query = query.Where("platform = ?", strings.ToLower(platform))
	}

	var analyses []models.InfluencerAnalysis
	if err := query.Order("created_at DESC").Find(&analyses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analyses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analyses": analyses, "count": len(analyses)})
}

func (h *ScoreHandler) CreateAnalysis(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := utils.ValidateInfluencerName(req.InfluencerName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.ValidatePlatform(req.Platform); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.AnalysisType) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "analysis_type is required"})
		return
	}
	if strings.TrimSpace(req.Summary) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "summary is required"})
		return
	}

	if req.ScoreID != nil {
		var score models.InfluencerScore
		if err := h.db.Where("id = ? AND user_id = ?", *req.ScoreID, userID).First(&score).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "linked score_id not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify score"})
			return
		}
	}

	analysis := models.InfluencerAnalysis{
		UserID:         userID,
		InfluencerName: strings.TrimSpace(req.InfluencerName),
		Platform:       strings.ToLower(strings.TrimSpace(req.Platform)),
		AnalysisType:   strings.TrimSpace(req.AnalysisType),
		Summary:        req.Summary,
		Insights:       req.Insights,
		RawLLMOutput:   req.RawLLMOutput,
		ScoreID:        req.ScoreID,
	}

	if err := h.db.Create(&analysis).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create analysis"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"analysis": analysis})
}

func validateScoreInput(name, platform string, scores ...float64) error {
	if err := utils.ValidateInfluencerName(name); err != nil {
		return err
	}
	if err := utils.ValidatePlatform(platform); err != nil {
		return err
	}
	for _, s := range scores {
		if s != 0 {
			if err := utils.ValidateScore(s); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseIDParam(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}
