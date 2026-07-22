package websocket

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	apimgmtEvent "github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/event"
	iamEvent "github.com/masterfabric-go/masterfabric/internal/domain/iam/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/realtime/model"
	tenantEvent "github.com/masterfabric-go/masterfabric/internal/domain/tenant/event"
	"github.com/masterfabric-go/masterfabric/internal/domain/tenant/repository"
	"github.com/masterfabric-go/masterfabric/internal/shared/events"
)

// EventBridge fans out domain events from the event bus to WebSocket rooms.
type EventBridge struct {
	hub     *Hub
	appRepo repository.AppRepository
	logger  *slog.Logger
}

// NewEventBridge creates a new EventBridge.
func NewEventBridge(hub *Hub, appRepo repository.AppRepository, logger *slog.Logger) *EventBridge {
	return &EventBridge{hub: hub, appRepo: appRepo, logger: logger}
}

// Register subscribes the bridge to all platform event topics.
func (b *EventBridge) Register(bus events.EventBus) {
	bus.Subscribe(events.TopicIAM, b.handleEvent(events.TopicIAM, "iam"))
	bus.Subscribe(events.TopicTenant, b.handleEvent(events.TopicTenant, "tenant"))
	bus.Subscribe(events.TopicAPIManagement, b.handleEvent(events.TopicAPIManagement, "api-management"))
}

func (b *EventBridge) handleEvent(topic, channel string) events.Handler {
	return func(ctx context.Context, event events.Event) error {
		routes := extractRoutes(event)
		if len(routes) == 0 {
			return nil
		}

		for _, route := range routes {
			orgID := route.orgID
			appID := route.appID

			if appID != uuid.Nil && orgID == uuid.Nil && b.appRepo != nil {
				app, err := b.appRepo.GetByID(ctx, appID)
				if err != nil {
					continue
				}
				orgID = app.OrganizationID
			}

			if orgID == uuid.Nil {
				continue
			}

			data, err := json.Marshal(event)
			if err != nil {
				if b.logger != nil {
					b.logger.Error("failed to marshal event for websocket", "error", err, "type", route.eventType)
				}
				continue
			}

			msg := model.NewEventMessage(route.eventType, topic, orgID.String(), appID.String(), data)
			payload, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			channels := []string{model.DefaultChannel, channel}
			seen := make(map[string]struct{})
			for _, ch := range channels {
				if _, dup := seen[ch]; dup {
					continue
				}
				seen[ch] = struct{}{}

				if appID != uuid.Nil {
					room, err := model.BuildRoomKey(orgID, appID, ch)
					if err != nil {
						continue
					}
					b.hub.Broadcast(room, payload)
				} else {
					b.hub.BroadcastToOrganization(orgID, ch, payload)
				}
			}
		}
		return nil
	}
}

type eventRoute struct {
	orgID     uuid.UUID
	appID     uuid.UUID
	eventType string
}

func extractRoutes(event events.Event) []eventRoute {
	switch e := event.(type) {
	case iamEvent.UserRegistered:
		return nil // no org/app scope
	case iamEvent.RoleAssigned:
		if e.OrganizationID == uuid.Nil {
			return nil
		}
		return []eventRoute{{
			orgID:     e.OrganizationID,
			appID:     uuid.Nil,
			eventType: "role.assigned",
		}}
	case tenantEvent.OrganizationCreated:
		return []eventRoute{{
			orgID:     e.OrganizationID,
			appID:     uuid.Nil,
			eventType: "organization.created",
		}}
	case tenantEvent.AppCreated:
		return []eventRoute{{
			orgID:     e.OrganizationID,
			appID:     e.AppID,
			eventType: "app.created",
		}}
	case tenantEvent.AppUpdated:
		return []eventRoute{{
			orgID:     e.OrganizationID,
			appID:     e.AppID,
			eventType: "app.updated",
		}}
	case tenantEvent.WorkspaceCreated:
		return []eventRoute{{
			orgID:     e.OrganizationID,
			appID:     uuid.Nil,
			eventType: "workspace.created",
		}}
	case apimgmtEvent.EndpointCreated:
		return []eventRoute{{
			orgID:     uuid.Nil,
			appID:     e.AppID,
			eventType: "endpoint.created",
		}}
	case apimgmtEvent.EndpointUpdated:
		return []eventRoute{{
			orgID:     uuid.Nil,
			appID:     e.AppID,
			eventType: "endpoint.updated",
		}}
	case apimgmtEvent.EndpointRetired:
		return []eventRoute{{
			orgID:     uuid.Nil,
			appID:     e.AppID,
			eventType: "endpoint.retired",
		}}
	case apimgmtEvent.EndpointActivated:
		return []eventRoute{{
			orgID:     uuid.Nil,
			appID:     e.AppID,
			eventType: "endpoint.activated",
		}}
	default:
		return nil
	}
}
