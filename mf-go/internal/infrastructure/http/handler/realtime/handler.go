// Package realtime provides the WebSocket HTTP handler for real-time event delivery.
package realtime

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"
	realtimeUC "github.com/masterfabric-go/masterfabric/internal/application/realtime/usecase"
	iamService "github.com/masterfabric-go/masterfabric/internal/domain/iam/service"
	"github.com/masterfabric-go/masterfabric/internal/domain/realtime/model"
	realtimeService "github.com/masterfabric-go/masterfabric/internal/domain/realtime/service"
	infraWS "github.com/masterfabric-go/masterfabric/internal/infrastructure/websocket"
	"github.com/masterfabric-go/masterfabric/internal/shared/middleware"
	"github.com/masterfabric-go/masterfabric/internal/shared/response"
)

// Handler handles WebSocket upgrade requests.
type Handler struct {
	validateUC *realtimeUC.ValidateConnectUseCase
	auth       iamService.AuthService
	hub        *infraWS.Hub
	upgrader   gorillaws.Upgrader
	pingSecs   int
	logger     *slog.Logger
	enabled    bool
}

// Config holds handler dependencies.
type Config struct {
	ValidateUC     *realtimeUC.ValidateConnectUseCase
	AuthService    iamService.AuthService
	Hub            *infraWS.Hub
	Upgrader       gorillaws.Upgrader
	PingInterval   int
	Logger         *slog.Logger
	Enabled        bool
}

// NewHandler creates a new WebSocket handler.
func NewHandler(cfg Config) *Handler {
	return &Handler{
		validateUC: cfg.ValidateUC,
		auth:       cfg.AuthService,
		hub:        cfg.Hub,
		upgrader:   cfg.Upgrader,
		pingSecs:   cfg.PingInterval,
		logger:     cfg.Logger,
		enabled:    cfg.Enabled,
	}
}

// Connect upgrades the HTTP connection to WebSocket and registers the client.
func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{"error": "websocket is disabled"})
		return
	}

	token := middleware.ExtractBearerToken(r)
	if token == "" {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
		return
	}

	claims, err := h.auth.ValidateToken(r.Context(), token)
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	orgID := resolveOrgID(r, claims.OrganizationID)
	appIDStr := r.Header.Get("X-App-ID")
	appID, err := realtimeUC.ParseAppHeader(appIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	input, err := h.validateUC.Execute(r.Context(), claims.UserID, orgID, appID)
	if err != nil {
		response.Error(w, err)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("websocket upgrade failed", "error", err)
		}
		return
	}

	clientID := infraWS.NewClientID()
	send := make(chan []byte, 64)
	info := realtimeService.ClientInfo{
		ID:             clientID,
		UserID:         input.UserID,
		OrganizationID: input.OrganizationID,
		AppID:          input.AppID,
	}

	unregister := h.hub.Register(info, send)

	wsClient := &infraWS.Session{
		ID:       clientID,
		Conn:     conn,
		Send:     send,
		Hub:      h.hub,
		OnAction: h.handleClientAction,
	}
	wsClient.Start(infraWS.PingInterval(h.pingSecs), unregister)

	if h.logger != nil {
		h.logger.Info("websocket client connected",
			"client_id", clientID,
			"user_id", input.UserID,
			"org_id", input.OrganizationID,
			"app_id", input.AppID,
		)
	}

	// Send welcome subscribed message for default channel.
	welcome, _ := json.Marshal(model.NewControlMessage(model.TypeSubscribed, model.DefaultChannel, "connected"))
	select {
	case send <- welcome:
	default:
	}
}

func (h *Handler) handleClientAction(clientID string, raw []byte) {
	var msg model.InboundMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		h.sendError(clientID, "invalid message format")
		return
	}

	switch msg.Action {
	case model.ActionPing:
		h.sendControl(clientID, model.TypePong, "", "")
	case model.ActionSubscribe:
		if err := model.ValidateChannelName(msg.Channel); err != nil {
			h.sendError(clientID, err.Error())
			return
		}
		if err := h.hub.Subscribe(clientID, msg.Channel); err != nil {
			h.sendError(clientID, err.Error())
			return
		}
		h.sendControl(clientID, model.TypeSubscribed, msg.Channel, "")
	case model.ActionUnsubscribe:
		if err := model.ValidateChannelName(msg.Channel); err != nil {
			h.sendError(clientID, err.Error())
			return
		}
		if err := h.hub.Unsubscribe(clientID, msg.Channel); err != nil {
			h.sendError(clientID, err.Error())
			return
		}
	default:
		h.sendError(clientID, "unknown action")
	}
}

func (h *Handler) sendControl(clientID, msgType, channel, message string) {
	payload, _ := json.Marshal(model.NewControlMessage(msgType, channel, message))
	h.hub.SendToClient(clientID, payload)
}

func (h *Handler) sendError(clientID, message string) {
	h.sendControl(clientID, model.TypeError, "", message)
}

func resolveOrgID(r *http.Request, claimOrgID uuid.UUID) uuid.UUID {
	if header := r.Header.Get("X-Organization-ID"); header != "" {
		if parsed, err := uuid.Parse(header); err == nil {
			return parsed
		}
	}
	if claimOrgID != uuid.Nil {
		return claimOrgID
	}
	if tenantOrg, ok := middleware.TenantIDFromContext(r.Context()); ok {
		return tenantOrg
	}
	return uuid.Nil
}
