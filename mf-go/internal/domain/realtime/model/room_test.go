package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRoomKey(t *testing.T) {
	orgID := uuid.New()
	appID := uuid.New()

	key, err := BuildRoomKey(orgID, appID, "events")
	require.NoError(t, err)
	assert.Contains(t, string(key), orgID.String())
	assert.Contains(t, string(key), appID.String())
	assert.Contains(t, string(key), "channel:events")
}

func TestBuildRoomKey_InvalidChannel(t *testing.T) {
	_, err := BuildRoomKey(uuid.New(), uuid.New(), "../bad")
	assert.Error(t, err)
}

func TestParseRoomKey_RoundTrip(t *testing.T) {
	orgID := uuid.New()
	appID := uuid.New()
	channel := "api-management"

	key, err := BuildRoomKey(orgID, appID, channel)
	require.NoError(t, err)

	parsedOrg, parsedApp, parsedChannel, err := ParseRoomKey(key)
	require.NoError(t, err)
	assert.Equal(t, orgID, parsedOrg)
	assert.Equal(t, appID, parsedApp)
	assert.Equal(t, channel, parsedChannel)
}

func TestValidateChannelName(t *testing.T) {
	assert.NoError(t, ValidateChannelName("events"))
	assert.NoError(t, ValidateChannelName("api-management"))
	assert.Error(t, ValidateChannelName("bad channel"))
	assert.Error(t, ValidateChannelName(""))
}
