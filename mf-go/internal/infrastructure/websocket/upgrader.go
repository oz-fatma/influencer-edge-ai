package websocket

import (
	"net/http"
	"strings"
	"time"

	gorillaws "github.com/gorilla/websocket"
)

const (
	defaultReadBufferSize  = 1024
	defaultWriteBufferSize = 1024
	defaultPingInterval    = 30 * time.Second
)

// UpgraderConfig holds WebSocket upgrader settings.
type UpgraderConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	AllowedOrigins  []string
}

// NewUpgrader creates a configured gorilla/websocket upgrader.
func NewUpgrader(cfg UpgraderConfig) gorillaws.Upgrader {
	readBuf := cfg.ReadBufferSize
	if readBuf <= 0 {
		readBuf = defaultReadBufferSize
	}
	writeBuf := cfg.WriteBufferSize
	if writeBuf <= 0 {
		writeBuf = defaultWriteBufferSize
	}

	return gorillaws.Upgrader{
		ReadBufferSize:  readBuf,
		WriteBufferSize: writeBuf,
		CheckOrigin:     originChecker(cfg.AllowedOrigins),
	}
}

func originChecker(allowed []string) func(*http.Request) bool {
	if len(allowed) == 0 {
		return func(*http.Request) bool { return true }
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		allowedSet[strings.TrimSpace(o)] = struct{}{}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		if _, ok := allowedSet["*"]; ok {
			return true
		}
		_, ok := allowedSet[origin]
		return ok
	}
}

// PingInterval returns the configured ping interval.
func PingInterval(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultPingInterval
	}
	return time.Duration(seconds) * time.Second
}
