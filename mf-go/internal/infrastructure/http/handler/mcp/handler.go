package mcp

import (
	"encoding/json"
	"net/http"
	"strings"

	infraMCP "github.com/masterfabric-go/masterfabric/internal/infrastructure/mcp"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
)

// Handler exposes MCP HTTP endpoints.
type Handler struct {
	service *infraMCP.Service
}

func NewHandler(service *infraMCP.Service) *Handler {
	return &Handler{service: service}
}

// ProcessRequest handles POST /api/v1/mcp/request.
func (h *Handler) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if h.service == nil {
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "MCP service not configured",
		})
		return
	}

	var req infraMCP.MCPPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	result, err := h.service.HandleMCPRequest(r.Context(), req)
	if err != nil {
		writeMCPError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func writeMCPError(w http.ResponseWriter, err error) {
	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "not configured"):
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{"error": msg})
	case strings.Contains(lower, "required"),
		strings.Contains(lower, "unsupported request_type"),
		strings.Contains(lower, "context.influencer_name"),
		strings.Contains(lower, "context.platform"),
		strings.Contains(lower, "notes must be"):
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": msg})
	default:
		response.JSON(w, http.StatusBadGateway, map[string]string{"error": msg})
	}
}
