package influencer

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/influencer/usecase"
	"github.com/masterfabric-go/masterfabric/internal/infrastructure/llm"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
	"github.com/masterfabric-go/masterfabric/internal/shared/validator"
)

type Handler struct {
	scores     *usecase.ScoreService
	analyses   *usecase.AnalysisService
	monitoring *usecase.MonitoringService
	llm        *llm.Analyzer
}

func NewHandler(
	scores *usecase.ScoreService,
	analyses *usecase.AnalysisService,
	monitoring *usecase.MonitoringService,
	llmAnalyzer *llm.Analyzer,
) *Handler {
	return &Handler{scores: scores, analyses: analyses, monitoring: monitoring, llm: llmAnalyzer}
}

func (h *Handler) CreateScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.CreateScoreRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	score, err := h.scores.Create(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Created(w, map[string]any{"score": score})
}

func (h *Handler) ListScores(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	scores, err := h.scores.List(r.Context(), userID, r.URL.Query().Get("platform"), parseQueryInt(r, "limit"), parseQueryInt(r, "offset"))
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"scores": scores, "count": len(scores)})
}

func (h *Handler) GetScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid score id"})
		return
	}
	score, err := h.scores.Get(r.Context(), userID, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"score": score})
}

func (h *Handler) UpdateScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid score id"})
		return
	}
	var req dto.UpdateScoreRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	score, err := h.scores.Update(r.Context(), userID, id, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"score": score})
}

func (h *Handler) DeleteScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid score id"})
		return
	}
	if err := h.scores.Delete(r.Context(), userID, id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"message": "score deleted successfully"})
}

func (h *Handler) CreateAnalysis(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.CreateAnalysisRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	analysis, err := h.analyses.Create(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Created(w, map[string]any{"analysis": analysis})
}

func (h *Handler) ListAnalyses(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	analyses, err := h.analyses.List(r.Context(), userID, r.URL.Query().Get("platform"), parseQueryInt(r, "limit"), parseQueryInt(r, "offset"))
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"analyses": analyses, "count": len(analyses)})
}

func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid analysis id"})
		return
	}
	analysis, err := h.analyses.Get(r.Context(), userID, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"analysis": analysis})
}

func (h *Handler) AnalyzeInfluencerLLM(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.llm == nil {
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "LLM service not configured (set LLM_BASE_URL on the server)",
		})
		return
	}

	var req dto.AnalyzeInfluencerRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := dto.ValidateInfluencerName(req.InfluencerName); err != nil {
		response.Error(w, err)
		return
	}
	if err := dto.ValidatePlatform(req.Platform); err != nil {
		response.Error(w, err)
		return
	}
	if len(req.Notes) > dto.MaxNotesLen {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "notes must be at most 4096 characters"})
		return
	}

	result, rawOutput, err := h.llm.Analyze(r.Context(), req.InfluencerName, req.Platform, req.Notes)
	if err != nil {
		response.JSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, dto.AnalyzeInfluencerResponse{
		Result: dto.AnalyzeInfluencerResult{
			OverallScore:    result.OverallScore,
			EngagementScore: result.EngagementScore,
			AudienceScore:   result.AudienceScore,
			BrandFitScore:   result.BrandFitScore,
			Summary:         result.Summary,
			Insights:        result.Insights,
		},
		RawOutput: rawOutput,
	})
}

func (h *Handler) RecordLLMMetric(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req dto.RecordLLMMetricRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	metric, err := h.monitoring.Record(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Created(w, map[string]any{"message": "metric recorded", "metric": metric})
}

func (h *Handler) GetMonitoringStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	stats, err := h.monitoring.Stats(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

func parseUUIDParam(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func parseQueryInt(r *http.Request, key string) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return n
}
