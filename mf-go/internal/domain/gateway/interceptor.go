package gateway

import (
	"context"
	"net/http"
)

// Interceptor defines the interface for request/response interceptors.
type Interceptor interface {
	// InterceptRequest is called before the request is forwarded to the backend.
	// It can modify the request, validate it, or return an error to stop processing.
	InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error)

	// InterceptResponse is called after the backend response is received.
	// It can modify the response, transform it, or add headers.
	InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error)
}

// RequestTransformer transforms request headers, body, or query parameters.
type RequestTransformer interface {
	TransformRequest(ctx context.Context, req *http.Request) (*http.Request, error)
}

// ResponseTransformer transforms response headers, body, or status code.
type ResponseTransformer interface {
	TransformResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error)
}

// RequestValidator validates request schema, headers, or body.
type RequestValidator interface {
	ValidateRequest(ctx context.Context, req *http.Request) error
}

// ResponseCache provides caching for responses.
type ResponseCache interface {
	Get(ctx context.Context, key string) (*http.Response, bool)
	Set(ctx context.Context, key string, resp *http.Response, ttl int) error
}

// Chain chains multiple interceptors together.
type Chain struct {
	interceptors []Interceptor
}

// NewChain creates a new interceptor chain.
func NewChain(interceptors ...Interceptor) *Chain {
	return &Chain{interceptors: interceptors}
}

// InterceptRequest runs all request interceptors in order.
func (c *Chain) InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	currentReq := req
	for _, interceptor := range c.interceptors {
		var err error
		currentReq, err = interceptor.InterceptRequest(ctx, currentReq)
		if err != nil {
			return nil, err
		}
	}
	return currentReq, nil
}

// InterceptResponse runs all response interceptors in reverse order.
func (c *Chain) InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error) {
	currentResp := resp
	for i := len(c.interceptors) - 1; i >= 0; i-- {
		var err error
		currentResp, err = c.interceptors[i].InterceptResponse(ctx, req, currentResp)
		if err != nil {
			return nil, err
		}
	}
	return currentResp, nil
}
