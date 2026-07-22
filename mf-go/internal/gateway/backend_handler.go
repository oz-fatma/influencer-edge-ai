package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
)

// BackendHandler defines the interface for backend service handlers.
// Each backend service (e.g., "product-service", "order-service") implements this interface.
type BackendHandler interface {
	// Handle processes the request and returns a response.
	// The endpoint contains metadata about the endpoint (method, path, schema, etc.)
	// The request has already been validated by the gateway pipeline.
	Handle(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error)
}

// BackendRegistry manages backend service handlers.
// It routes requests to the appropriate handler based on backend_service name.
// Implements HandlerResolver interface for compatibility with dynamic resolution.
type BackendRegistry struct {
	handlers map[string]BackendHandler
}

// NewBackendRegistry creates a new backend registry.
func NewBackendRegistry() *BackendRegistry {
	return &BackendRegistry{
		handlers: make(map[string]BackendHandler),
	}
}

// Register registers a handler for a backend service.
// Example: registry.Register("product-service", productHandler)
func (r *BackendRegistry) Register(serviceName string, handler BackendHandler) {
	if r.handlers == nil {
		r.handlers = make(map[string]BackendHandler)
	}
	r.handlers[serviceName] = handler
}

// Handle routes the request to the appropriate backend handler.
func (r *BackendRegistry) Handle(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	handler, ok := r.handlers[endpoint.BackendService]
	if !ok {
		return nil, fmt.Errorf("backend service '%s' not registered", endpoint.BackendService)
	}
	return handler.Handle(ctx, endpoint, req)
}

// GetHandler returns the handler for a service if registered, otherwise nil.
func (r *BackendRegistry) GetHandler(serviceName string) BackendHandler {
	return r.handlers[serviceName]
}

// IsRegistered checks if a backend service has a registered handler.
func (r *BackendRegistry) IsRegistered(serviceName string) bool {
	_, ok := r.handlers[serviceName]
	return ok
}

// ListRegisteredServices returns all registered backend service names.
func (r *BackendRegistry) ListRegisteredServices() []string {
	services := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		services = append(services, name)
	}
	return services
}
