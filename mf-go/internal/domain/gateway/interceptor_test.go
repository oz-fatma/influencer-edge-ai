package gateway

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockInterceptor is a test interceptor
type mockInterceptor struct {
	requestCalled  bool
	responseCalled bool
	requestError   error
	responseError  error
}

func (m *mockInterceptor) InterceptRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	m.requestCalled = true
	if m.requestError != nil {
		return nil, m.requestError
	}
	return req, nil
}

func (m *mockInterceptor) InterceptResponse(ctx context.Context, req *http.Request, resp *http.Response) (*http.Response, error) {
	m.responseCalled = true
	if m.responseError != nil {
		return nil, m.responseError
	}
	return resp, nil
}

func TestChain_InterceptRequest(t *testing.T) {
	interceptor1 := &mockInterceptor{}
	interceptor2 := &mockInterceptor{}

	chain := NewChain(interceptor1, interceptor2)

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)

	result, err := chain.InterceptRequest(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, interceptor1.requestCalled)
	assert.True(t, interceptor2.requestCalled)
}

func TestChain_InterceptRequest_Error(t *testing.T) {
	interceptor1 := &mockInterceptor{requestError: errors.New("test error")}
	interceptor2 := &mockInterceptor{}

	chain := NewChain(interceptor1, interceptor2)

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)

	result, err := chain.InterceptRequest(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, interceptor1.requestCalled)
	assert.False(t, interceptor2.requestCalled) // Should stop on error
}

func TestChain_InterceptResponse(t *testing.T) {
	interceptor1 := &mockInterceptor{}
	interceptor2 := &mockInterceptor{}

	chain := NewChain(interceptor1, interceptor2)

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp := &http.Response{StatusCode: http.StatusOK}

	result, err := chain.InterceptResponse(context.Background(), req, resp)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, interceptor1.responseCalled)
	assert.True(t, interceptor2.responseCalled)
}

func TestChain_InterceptResponse_ReverseOrder(t *testing.T) {
	interceptor1 := &mockInterceptor{}
	interceptor2 := &mockInterceptor{}

	chain := NewChain(interceptor1, interceptor2)

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp := &http.Response{StatusCode: http.StatusOK}

	_, err := chain.InterceptResponse(context.Background(), req, resp)
	require.NoError(t, err)
	
	// Both should be called
	assert.True(t, interceptor1.responseCalled)
	assert.True(t, interceptor2.responseCalled)
}
