package admin

import (
	"net/http"
	"strconv"

	"github.com/masterfabric-go/masterfabric/internal/application/admin/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/admin/usecase"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
	"github.com/masterfabric-go/masterfabric/internal/shared/validator"
)

type Handler struct {
	config *usecase.ConfigService
	logs   *usecase.LogsService
}

func NewHandler(config *usecase.ConfigService, logs *usecase.LogsService) *Handler {
	return &Handler{config: config, logs: logs}
}

func (h *Handler) GetLLMConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.config.Get(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, cfg)
}

func (h *Handler) UpdateLLMConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	var req dto.UpdateLLMConfigRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	cfg, err := h.config.Update(r.Context(), userID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, cfg)
}

func (h *Handler) ListLLMLogs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	logs, err := h.logs.ListRecent(r.Context(), limit)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, logs)
}

func (h *Handler) ListAllowedModels(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"models": dto.AllowedModels()})
}
