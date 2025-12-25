//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

const envValueTrue = "true"

// skipTrafficTestsIfNeeded skips traffic tests when host networking is not available.
// On macOS with Podman, containers cannot reach host.docker.internal.
func skipTrafficTestsIfNeeded(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_TRAFFIC_TESTS") == envValueTrue {
		t.Skip("Skipping traffic tests: SKIP_TRAFFIC_TESTS=true")
	}

	// On macOS, Podman doesn't support host.docker.internal well
	if runtime.GOOS == "darwin" && os.Getenv("FORCE_TRAFFIC_TESTS") != envValueTrue {
		t.Skip("Skipping traffic tests on macOS (host.docker.internal not supported with Podman). Set FORCE_TRAFFIC_TESTS=true to force.")
	}
}

// getContainerAccessibleAddress converts a host URL to an address accessible from inside the container.
func getContainerAccessibleAddress(hostURL string) string {
	u, err := url.Parse(hostURL)
	if err != nil {
		return hostURL
	}

	port := u.Port()
	if port == "" {
		port = "80"
	}

	// Docker Desktop on Mac/Windows uses host.docker.internal
	// Podman on Linux uses host.containers.internal
	hostGateway := "host.docker.internal"
	if runtime.GOOS == "linux" {
		hostGateway = "host.containers.internal"
	}

	return fmt.Sprintf("%s:%s", hostGateway, port)
}

// sendHTTPRequest sends an HTTP request through the proxy.
func sendHTTPRequest(ctx context.Context, proxyAddr, path, host string, headers map[string]string) (*http.Response, error) {
	reqURL := fmt.Sprintf("http://%s%s", proxyAddr, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Host = host

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		// Don't follow redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func TestTraffic_BasicProxy(t *testing.T) {
	t.Parallel()
	skipTrafficTestsIfNeeded(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start mock backend
	backend := StartMockBackend()
	defer backend.Close()

	// Start Pingora container
	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Configure route pointing to mock backend
	backendAddr := getContainerAccessibleAddress(backend.URL())

	route := NewHTTPRoute("default/test", []string{"test.example.com"}, "/", backendAddr)

	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{route},
		Version:    1,
	})
	require.NoError(t, err)
	require.True(t, resp.GetSuccess())

	// Send HTTP request through proxy
	httpResp, err := sendHTTPRequest(ctx, container.HTTPAddr, "/hello", "test.example.com", nil)
	require.NoError(t, err)
	defer httpResp.Body.Close()

	assert.Equal(t, http.StatusOK, httpResp.StatusCode)
	assert.Equal(t, "true", httpResp.Header.Get("X-Backend-Received"))

	// Verify backend received the request
	requests := backend.GetRequests()
	require.Len(t, requests, 1)
	assert.Equal(t, "/hello", requests[0].Path)
	assert.Equal(t, http.MethodGet, requests[0].Method)
}

func TestTraffic_PathPrefix(t *testing.T) {
	t.Parallel()
	skipTrafficTestsIfNeeded(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	backendAPI := StartMockBackend()
	defer backendAPI.Close()

	backendStatic := StartMockBackend()
	defer backendStatic.Close()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Configure routes with different path prefixes
	routes := []*routingv1.HTTPRoute{
		NewHTTPRoute("default/api", []string{"app.example.com"}, "/api", getContainerAccessibleAddress(backendAPI.URL())),
		NewHTTPRoute("default/static", []string{"app.example.com"}, "/static", getContainerAccessibleAddress(backendStatic.URL())),
	}

	_, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: routes,
		Version:    1,
	})
	require.NoError(t, err)

	// Test /api/users -> backendAPI
	resp1, err := sendHTTPRequest(ctx, container.HTTPAddr, "/api/users", "app.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp1.Body)
	resp1.Body.Close()

	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, 1, backendAPI.RequestCount())
	assert.Equal(t, 0, backendStatic.RequestCount())

	// Test /static/image.png -> backendStatic
	backendAPI.Reset()

	resp2, err := sendHTTPRequest(ctx, container.HTTPAddr, "/static/image.png", "app.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp2.Body)
	resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	assert.Equal(t, 0, backendAPI.RequestCount())
	assert.Equal(t, 1, backendStatic.RequestCount())
}

func TestTraffic_PathExact(t *testing.T) {
	t.Parallel()
	skipTrafficTestsIfNeeded(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	backend := StartMockBackend()
	defer backend.Close()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Configure route with exact path match
	route := NewHTTPRouteExact("default/exact", []string{"app.example.com"}, "/health", getContainerAccessibleAddress(backend.URL()))

	_, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{route},
		Version:    1,
	})
	require.NoError(t, err)

	// Exact match should work
	resp1, err := sendHTTPRequest(ctx, container.HTTPAddr, "/health", "app.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp1.Body)
	resp1.Body.Close()

	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, 1, backend.RequestCount())

	// Subpath should NOT match exact route
	backend.Reset()

	resp2, err := sendHTTPRequest(ctx, container.HTTPAddr, "/health/check", "app.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp2.Body)
	resp2.Body.Close()

	// Should get 404 or not reach backend
	assert.Equal(t, 0, backend.RequestCount(), "Exact match should not match subpaths")
}

func TestTraffic_HostRouting(t *testing.T) {
	t.Parallel()
	skipTrafficTestsIfNeeded(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	backendA := StartMockBackend()
	defer backendA.Close()

	backendB := StartMockBackend()
	defer backendB.Close()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Configure routes for different hosts
	routes := []*routingv1.HTTPRoute{
		NewHTTPRoute("default/app-a", []string{"a.example.com"}, "/", getContainerAccessibleAddress(backendA.URL())),
		NewHTTPRoute("default/app-b", []string{"b.example.com"}, "/", getContainerAccessibleAddress(backendB.URL())),
	}

	_, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: routes,
		Version:    1,
	})
	require.NoError(t, err)

	// Request to a.example.com -> backendA
	resp1, err := sendHTTPRequest(ctx, container.HTTPAddr, "/test", "a.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp1.Body)
	resp1.Body.Close()

	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, 1, backendA.RequestCount())
	assert.Equal(t, 0, backendB.RequestCount())

	// Request to b.example.com -> backendB
	backendA.Reset()

	resp2, err := sendHTTPRequest(ctx, container.HTTPAddr, "/test", "b.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp2.Body)
	resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	assert.Equal(t, 0, backendA.RequestCount())
	assert.Equal(t, 1, backendB.RequestCount())
}

func TestTraffic_NoRoute404(t *testing.T) {
	t.Parallel()
	skipTrafficTestsIfNeeded(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Configure route for specific host only
	backend := StartMockBackend()
	defer backend.Close()

	route := NewHTTPRoute("default/known", []string{"known.example.com"}, "/", getContainerAccessibleAddress(backend.URL()))

	_, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{route},
		Version:    1,
	})
	require.NoError(t, err)

	// Request to unknown host should fail
	resp, err := sendHTTPRequest(ctx, container.HTTPAddr, "/test", "unknown.example.com", nil)
	require.NoError(t, err)
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Expect 404 or 502 (no route found)
	assert.True(t, resp.StatusCode >= 400, "Expected 4xx or 5xx status, got %d", resp.StatusCode)
	assert.Equal(t, 0, backend.RequestCount(), "Backend should not receive request for unknown host")
}
