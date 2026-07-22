# Dynamic Backend Handler System

The Dynamic Backend Handler System automatically resolves and routes requests to backend services **without requiring pre-registration**. It supports multiple strategies for handling requests dynamically based on endpoint configuration.

## Architecture

1. **DynamicHandlerResolver**: Automatically resolves handlers using multiple strategies
2. **BackendRegistry**: Manages registered handlers (optional, for custom handlers)
3. **Gateway Pipeline**: Integrates the resolver to route validated requests dynamically

## How It Works

The system uses **three strategies** in order:

### Strategy 1: Registered Handlers (Optional)
If you register a specific handler for a service, it will be used.

### Strategy 2: HTTP Proxy to External Services
If `backend_service` is a URL (e.g., `https://api.example.com`) or has service configuration, requests are automatically proxied to that URL.

### Strategy 3: Generic Dynamic Handler
If no handler or proxy is configured, a generic handler processes the request based on endpoint metadata (`backend_action`, `method`, etc.).

**No pre-registration required!** Endpoints work immediately after being defined.

## Usage Examples

### Option 1: Use Generic Dynamic Handler (No Configuration Needed)

Simply define an endpoint with `backend_service` and `backend_action`:

```json
{
  "method": "GET",
  "path": "/products",
  "backend_service": "product-service",
  "backend_action": "list"
}
```

The gateway will automatically handle the request using the generic dynamic handler.

### Option 2: HTTP Proxy to External Service

#### Method A: Use URL as backend_service

```json
{
  "method": "GET",
  "path": "/products",
  "backend_service": "https://api.example.com/products",
  "backend_action": "list"
}
```

Requests will be automatically proxied to `https://api.example.com/products`.

#### Method B: Configure Service Mapping

In `cmd/server/main.go`:

```go
dynamicResolver.RegisterServiceConfig("product-service", gateway.ServiceConfig{
    BaseURL: "https://api.example.com/products",
    Headers: map[string]string{
        "Authorization": "Bearer your-token",
        "X-Custom-Header": "value",
    },
    Timeout: 30, // seconds
})
```

Then define endpoints with `backend_service: "product-service"`.

### Option 3: Register Custom Handler (For Complex Logic)

If you need custom business logic:

```go
// 1. Implement BackendHandler interface
type ProductHandler struct {
    productRepo repository.ProductRepository
}

func (h *ProductHandler) Handle(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
    switch endpoint.BackendAction {
    case "list":
        products, _ := h.productRepo.List(ctx)
        // ... return response
    }
}

// 2. Register in main.go
backendRegistry := gateway.NewBackendRegistry()
productHandler := handlers.NewProductHandler(productRepo)
backendRegistry.Register("product-service", productHandler)

// 3. Create dynamic resolver with registry
dynamicResolver := gateway.NewDynamicHandlerResolver(backendRegistry, log)
```

## Example Handler Implementation

See `internal/gateway/handlers/example_handler.go` for a complete example implementation.

## Defining Endpoints

When creating an endpoint via the API, specify:

```json
{
  "method": "GET",
  "path": "/products",
  "backend_service": "product-service",
  "backend_action": "list",
  "schema": { ... },
  "pii_masking": { ... }
}
```

## Response Format

Handlers should return `*http.Response` with:
- Appropriate status code
- `Content-Type: application/json` header
- JSON body in the response

Example:
```go
body, _ := json.Marshal(map[string]interface{}{
    "data": products,
    "count": len(products),
})

resp := &http.Response{
    StatusCode: http.StatusOK,
    Header:     make(http.Header),
    Body:       io.NopCloser(bytes.NewReader(body)),
}
resp.Header.Set("Content-Type", "application/json")
return resp, nil
```

## Benefits

1. **Zero Configuration**: Endpoints work immediately after being defined - no code changes needed
2. **Flexible Routing**: Supports registered handlers, HTTP proxying, and generic handling
3. **Dynamic**: Handlers are resolved at runtime based on endpoint configuration
4. **Backward Compatible**: Still supports registered handlers for custom logic
5. **Separation of Concerns**: Gateway handles policies, handlers handle business logic
6. **Scalability**: Different services can be implemented and deployed separately

## How Dynamic Resolution Works

When a request comes in:

1. **Gateway validates** the endpoint (permissions, rate limits, schema, etc.)
2. **Dynamic resolver** tries to resolve a handler:
   - Checks if a handler is registered → Use registered handler
   - Checks if `backend_service` is a URL → Proxy to that URL
   - Checks if service has configuration → Proxy using configuration
   - Otherwise → Use generic dynamic handler
3. **Handler processes** the request based on `backend_action` and `method`
4. **Response** is returned to the client

## Generic Dynamic Handler Behavior

The generic handler processes requests based on:
- **HTTP Method**: GET, POST, PUT, PATCH, DELETE
- **Backend Action**: list, get, create, update, delete
- **Endpoint Metadata**: path, schema, etc.

It returns appropriate responses for each action type, allowing endpoints to work immediately while you implement specific handlers or configure proxies.
