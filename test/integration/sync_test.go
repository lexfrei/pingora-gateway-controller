//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

func TestSync_SingleHTTPRoute(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Create single route
	route := NewHTTPRoute("default/test-route", []string{"test.example.com"}, "/", "backend:8080")

	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{route},
		Version:    1,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess(), "UpdateRoutes should succeed")
	assert.Equal(t, uint32(1), resp.GetHttpRouteCount())
	assert.Equal(t, uint64(1), resp.GetAppliedVersion())
}

func TestSync_MultipleHTTPRoutes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Create multiple routes
	routes := []*routingv1.HTTPRoute{
		NewHTTPRoute("default/route1", []string{"app1.example.com"}, "/api", "backend1:8080"),
		NewHTTPRoute("default/route2", []string{"app2.example.com"}, "/api", "backend2:8080"),
		NewHTTPRoute("default/route3", []string{"app3.example.com"}, "/", "backend3:8080"),
	}

	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: routes,
		Version:    1,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, uint32(3), resp.GetHttpRouteCount())
}

func TestSync_UpdateRoute(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Initial route
	route := NewHTTPRoute("default/test", []string{"test.example.com"}, "/v1", "backend-v1:8080")
	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{route},
		Version:    1,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())

	// Update route (change backend)
	route.Rules[0].Backends[0].Address = "backend-v2:8080"
	resp, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{route},
		Version:    2,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, uint64(2), resp.GetAppliedVersion())
}

func TestSync_DeleteRoutes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Create routes
	routes := []*routingv1.HTTPRoute{
		NewHTTPRoute("default/route1", []string{"app1.example.com"}, "/", "backend1:8080"),
		NewHTTPRoute("default/route2", []string{"app2.example.com"}, "/", "backend2:8080"),
	}
	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: routes,
		Version:    1,
	})
	require.NoError(t, err)
	assert.Equal(t, uint32(2), resp.GetHttpRouteCount())

	// Delete all routes by sending empty list
	resp, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: []*routingv1.HTTPRoute{},
		Version:    2,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, uint32(0), resp.GetHttpRouteCount())

	// Verify via GetRoutes
	getResp, err := client.GetRoutes(ctx, &routingv1.GetRoutesRequest{})
	require.NoError(t, err)
	assert.Empty(t, getResp.GetHttpRoutes())
}

func TestSync_GRPCRoute(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Create gRPC route
	route := NewGRPCRoute("default/grpc-test", []string{"grpc.example.com"}, "example.MyService", "DoSomething", "grpc-backend:9000")

	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		GrpcRoutes: []*routingv1.GRPCRoute{route},
		Version:    1,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, uint32(1), resp.GetGrpcRouteCount())
}

func TestSync_MixedRoutes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Create mixed HTTP and gRPC routes
	httpRoutes := []*routingv1.HTTPRoute{
		NewHTTPRoute("default/http1", []string{"api.example.com"}, "/v1", "http-backend:8080"),
		NewHTTPRoute("default/http2", []string{"web.example.com"}, "/", "web-backend:8080"),
	}
	grpcRoutes := []*routingv1.GRPCRoute{
		NewGRPCRoute("default/grpc1", []string{"grpc.example.com"}, "svc.A", "MethodA", "grpc-a:9000"),
	}

	resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: httpRoutes,
		GrpcRoutes: grpcRoutes,
		Version:    1,
	})
	require.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, uint32(2), resp.GetHttpRouteCount())
	assert.Equal(t, uint32(1), resp.GetGrpcRouteCount())
}

func TestSync_Version(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Test version tracking across multiple updates
	for i := uint64(1); i <= 5; i++ {
		resp, err := client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
			HttpRoutes: []*routingv1.HTTPRoute{
				NewHTTPRoute("default/test", []string{"test.example.com"}, "/", "backend:8080"),
			},
			Version: i,
		})
		require.NoError(t, err)
		assert.True(t, resp.GetSuccess())
		assert.Equal(t, i, resp.GetAppliedVersion(), "Version %d should be applied", i)
	}
}

func TestSync_GetRoutes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, container.WaitForReady(ctx, 30*time.Second))

	// Create routes
	httpRoutes := []*routingv1.HTTPRoute{
		NewHTTPRoute("default/route1", []string{"app1.example.com"}, "/api", "backend1:8080"),
		NewHTTPRoute("default/route2", []string{"app2.example.com"}, "/", "backend2:8080"),
	}
	grpcRoutes := []*routingv1.GRPCRoute{
		NewGRPCRoute("default/grpc1", []string{"grpc.example.com"}, "svc.Test", "Call", "grpc:9000"),
	}

	_, err = client.UpdateRoutes(ctx, &routingv1.UpdateRoutesRequest{
		HttpRoutes: httpRoutes,
		GrpcRoutes: grpcRoutes,
		Version:    1,
	})
	require.NoError(t, err)

	// Retrieve and verify
	getResp, err := client.GetRoutes(ctx, &routingv1.GetRoutesRequest{})
	require.NoError(t, err)
	assert.Len(t, getResp.GetHttpRoutes(), 2)
	assert.Len(t, getResp.GetGrpcRoutes(), 1)

	// Verify route IDs
	routeIDs := make(map[string]bool)
	for _, r := range getResp.GetHttpRoutes() {
		routeIDs[r.GetId()] = true
	}

	assert.True(t, routeIDs["default/route1"])
	assert.True(t, routeIDs["default/route2"])
}
