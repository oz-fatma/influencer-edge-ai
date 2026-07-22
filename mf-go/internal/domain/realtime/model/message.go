package model

import (
	"encoding/json"
	"time"
)

// Client actions.
const (
	ActionSubscribe   = "subscribe"
	ActionUnsubscribe = "unsubscribe"
	ActionPing        = "ping"
)

// Server message types.
const (
	TypePong       = "pong"
	TypeSubscribed = "subscribed"
	TypeError      = "error"
)

// InboundMessage is a JSON message from the client.
type InboundMessage struct {
	Action  string `json:"action"`
	Channel string `json:"channel,omitempty"`
}

// OutboundMessage is a JSON message pushed to the client.
type OutboundMessage struct {
	Type           string          `json:"type"`
	Topic          string          `json:"topic,omitempty"`
	OrganizationID string          `json:"organization_id,omitempty"`
	AppID          string          `json:"app_id,omitempty"`
	Channel        string          `json:"channel,omitempty"`
	Message        string          `json:"message,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	Timestamp      time.Time       `json:"timestamp,omitempty"`
}

// NewEventMessage builds a domain event push payload.
func NewEventMessage(eventType, topic, orgID, appID string, data json.RawMessage) OutboundMessage {
	return OutboundMessage{
		Type:           eventType,
		Topic:          topic,
		OrganizationID: orgID,
		AppID:          appID,
		Data:           data,
		Timestamp:      time.Now().UTC(),
	}
}

// NewControlMessage builds a control response (pong, subscribed, error).
func NewControlMessage(msgType, channel, message string) OutboundMessage {
	return OutboundMessage{
		Type:      msgType,
		Channel:   channel,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
}
