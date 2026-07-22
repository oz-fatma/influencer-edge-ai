package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultChannel = "events"
	MaxChannelLen  = 64
)

var channelNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// RoomKey identifies a scoped broadcast room.
type RoomKey string

// BuildRoomKey creates a room key for org/app/channel scope.
func BuildRoomKey(orgID, appID uuid.UUID, channel string) (RoomKey, error) {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		channel = DefaultChannel
	}
	if !channelNameRe.MatchString(channel) {
		return "", fmt.Errorf("invalid channel name: %s", channel)
	}
	return RoomKey(fmt.Sprintf("org:%s:app:%s:channel:%s", orgID, appID, channel)), nil
}

// ParseRoomKey splits a room key into its components.
func ParseRoomKey(key RoomKey) (orgID, appID uuid.UUID, channel string, err error) {
	s := string(key)
	if !strings.HasPrefix(s, "org:") || !strings.Contains(s, ":app:") || !strings.Contains(s, ":channel:") {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid room key format")
	}

	orgAndRest := strings.TrimPrefix(s, "org:")
	appSplit := strings.SplitN(orgAndRest, ":app:", 2)
	if len(appSplit) != 2 {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid room key format")
	}

	channelSplit := strings.SplitN(appSplit[1], ":channel:", 2)
	if len(channelSplit) != 2 {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid room key format")
	}

	orgID, err = uuid.Parse(appSplit[0])
	if err != nil {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid organization id in room key")
	}
	appID, err = uuid.Parse(channelSplit[0])
	if err != nil {
		return uuid.Nil, uuid.Nil, "", fmt.Errorf("invalid app id in room key")
	}
	return orgID, appID, channelSplit[1], nil
}

// ValidateChannelName checks whether a channel name is allowed.
func ValidateChannelName(name string) error {
	if !channelNameRe.MatchString(strings.TrimSpace(name)) {
		return fmt.Errorf("channel name must be 1-64 alphanumeric, underscore, or hyphen characters")
	}
	return nil
}
