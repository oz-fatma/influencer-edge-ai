package websocket

import (
	"context"
	"testing"

	"github.com/google/uuid"
	realtimeService "github.com/masterfabric-go/masterfabric/internal/domain/realtime/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_RegisterBroadcastSubscribe(t *testing.T) {
	hub := NewHub(nil, 10)
	orgID := uuid.New()
	appID := uuid.New()

	send := make(chan []byte, 4)
	info := realtimeService.ClientInfo{
		ID:             "client-1",
		UserID:         uuid.New(),
		OrganizationID: orgID,
		AppID:          appID,
	}
	unregister := hub.Register(info, send)
	defer unregister()

	assert.Equal(t, 1, hub.ConnectionCount())

	require.NoError(t, hub.Subscribe("client-1", "tenant"))

	payload := []byte(`{"type":"test"}`)
	hub.BroadcastToOrganization(orgID, "events", payload)

	select {
	case msg := <-send:
		assert.Equal(t, payload, msg)
	default:
		t.Fatal("expected broadcast message")
	}
}

func TestHub_Close(t *testing.T) {
	hub := NewHub(nil, 10)
	send := make(chan []byte, 1)
	hub.Register(realtimeService.ClientInfo{
		ID:             "c1",
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		AppID:          uuid.New(),
	}, send)

	err := hub.Close(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, hub.ConnectionCount())
}

func TestHub_SendToClient(t *testing.T) {
	hub := NewHub(nil, 10)
	send := make(chan []byte, 1)
	hub.Register(realtimeService.ClientInfo{
		ID:             "c1",
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		AppID:          uuid.New(),
	}, send)

	hub.SendToClient("c1", []byte("pong"))
	assert.Equal(t, []byte("pong"), <-send)
}
