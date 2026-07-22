package events

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEvent struct {
	Message string
}

func TestInProcessBus_Publish(t *testing.T) {
	logger := slog.Default()
	bus := NewInProcessBus(logger, 10)

	var called int32
	bus.Subscribe("test.topic", func(ctx context.Context, event Event) error {
		e := event.(testEvent)
		assert.Equal(t, "hello", e.Message)
		atomic.AddInt32(&called, 1)
		return nil
	})

	err := bus.Publish(context.Background(), "test.topic", testEvent{Message: "hello"})
	require.NoError(t, err)

	// Wait for async dispatch
	time.Sleep(50 * time.Millisecond)
	_ = bus.Close()

	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestInProcessBus_MultipleSubscribers(t *testing.T) {
	logger := slog.Default()
	bus := NewInProcessBus(logger, 10)

	var count int32
	for i := 0; i < 3; i++ {
		bus.Subscribe("test.topic", func(ctx context.Context, event Event) error {
			atomic.AddInt32(&count, 1)
			return nil
		})
	}

	err := bus.Publish(context.Background(), "test.topic", testEvent{Message: "multi"})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	_ = bus.Close()

	assert.Equal(t, int32(3), atomic.LoadInt32(&count))
}

func TestInProcessBus_ImplementsInterface(t *testing.T) {
	logger := slog.Default()
	var bus EventBus = NewInProcessBus(logger, 10)
	assert.NotNil(t, bus)
	_ = bus.Close()
}

func TestNewEnvelope(t *testing.T) {
	data := testEvent{Message: "test"}
	env, err := NewEnvelope("test.event", "test-service", data)
	require.NoError(t, err)
	assert.Equal(t, "test.event", env.Type)
	assert.Equal(t, "test-service", env.Source)
	assert.NotEmpty(t, env.ID)
	assert.NotEmpty(t, env.Data)
}
