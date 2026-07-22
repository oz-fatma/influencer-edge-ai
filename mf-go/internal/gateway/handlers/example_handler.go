package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/masterfabric-go/masterfabric/internal/domain/apimanagement/model"
)

// ExampleProductHandler demonstrates how to implement a backend handler.
// This handler processes requests for the "product-service" backend.
type ExampleProductHandler struct {
	// Add your dependencies here (repositories, services, etc.)
	// productRepo repository.ProductRepository
	// logger *slog.Logger
}

// NewExampleProductHandler creates a new product handler.
func NewExampleProductHandler() *ExampleProductHandler {
	return &ExampleProductHandler{
		// Initialize dependencies
	}
}

// Handle processes the request based on the endpoint's backend_action.
func (h *ExampleProductHandler) Handle(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	// Route to specific action based on endpoint.BackendAction
	switch endpoint.BackendAction {
	case "list":
		return h.handleList(ctx, endpoint, req)
	case "get":
		return h.handleGet(ctx, endpoint, req)
	case "create":
		return h.handleCreate(ctx, endpoint, req)
	case "update":
		return h.handleUpdate(ctx, endpoint, req)
	case "delete":
		return h.handleDelete(ctx, endpoint, req)
	default:
		return h.handleDefault(ctx, endpoint, req)
	}
}

// handleList handles GET requests to list products.
func (h *ExampleProductHandler) handleList(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	// Example: Fetch products from repository
	// products, err := h.productRepo.List(ctx, ...)
	// if err != nil {
	//     return nil, err
	// }

	// For demonstration, return mock data
	products := []map[string]interface{}{
		{"id": "1", "name": "Product 1", "price": 99.99},
		{"id": "2", "name": "Product 2", "price": 149.99},
	}

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
}

// handleGet handles GET requests to get a single product.
func (h *ExampleProductHandler) handleGet(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	// Extract ID from path (e.g., /products/{id})
	// id := chi.URLParam(req, "id")
	// product, err := h.productRepo.GetByID(ctx, id)

	// For demonstration
	product := map[string]interface{}{
		"id":    "1",
		"name":  "Product 1",
		"price": 99.99,
	}

	body, _ := json.Marshal(map[string]interface{}{"data": product})

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// handleCreate handles POST requests to create a product.
func (h *ExampleProductHandler) handleCreate(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	// Parse request body
	var input map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"error":"invalid request body"}`)),
		}
		resp.Header.Set("Content-Type", "application/json")
		return resp, nil
	}

	// Create product
	// product, err := h.productRepo.Create(ctx, input)
	// if err != nil {
	//     return nil, err
	// }

	// For demonstration
	product := map[string]interface{}{
		"id":    "new-id",
		"name":  input["name"],
		"price": input["price"],
	}

	body, _ := json.Marshal(map[string]interface{}{"data": product})

	resp := &http.Response{
		StatusCode: http.StatusCreated,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// handleUpdate handles PUT/PATCH requests to update a product.
func (h *ExampleProductHandler) handleUpdate(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	// Similar to handleCreate
	body, _ := json.Marshal(map[string]interface{}{
		"message": "Product updated successfully",
	})

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// handleDelete handles DELETE requests to delete a product.
func (h *ExampleProductHandler) handleDelete(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"message": "Product deleted successfully",
	})

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// handleDefault handles requests with unknown actions.
func (h *ExampleProductHandler) handleDefault(ctx context.Context, endpoint *model.Endpoint, req *http.Request) (*http.Response, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"error": fmt.Sprintf("Unknown backend action: %s", endpoint.BackendAction),
	})

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}
