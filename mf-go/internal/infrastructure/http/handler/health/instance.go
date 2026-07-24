package health

import (
	"os"
	"strings"
)

// ResolveInstanceID returns the identifier for this running instance.
// INSTANCE_ID overrides HOSTNAME (Docker sets a unique hostname per container).
func ResolveInstanceID() string {
	if id := strings.TrimSpace(os.Getenv("INSTANCE_ID")); id != "" {
		return id
	}
	if host := strings.TrimSpace(os.Getenv("HOSTNAME")); host != "" {
		return host
	}
	return "unknown"
}
