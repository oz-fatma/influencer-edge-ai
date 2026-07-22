package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/domain/realtime/model"
)

// ClientInfo holds metadata for a connected WebSocket client.
type ClientInfo struct {
	ID             string
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	AppID          uuid.UUID
}

// Hub manages WebSocket client connections and room subscriptions.
type Hub interface {
	// Register adds a client and subscribes it to the default events channel.
	Register(client ClientInfo, send chan []byte) (unregister func())
	// Subscribe adds a client to an additional channel room.
	Subscribe(clientID string, channel string) error
	// Unsubscribe removes a client from a channel room.
	Unsubscribe(clientID string, channel string) error
	// Broadcast sends a message to all clients in a room.
	Broadcast(room model.RoomKey, payload []byte)
	// ConnectionCount returns the number of active connections.
	ConnectionCount() int
	// Close shuts down the hub.
	Close(ctx context.Context) error
}
