//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

// createGRPCClient creates a gRPC client connected to the Pingora proxy.
func createGRPCClient(_ context.Context, address string) (routingv1.RoutingServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	client := routingv1.NewRoutingServiceClient(conn)

	return client, conn, nil
}

func TestGRPC_Connection(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start Pingora container
	container, err := StartPingoraContainer(ctx)
	require.NoError(t, err, "Failed to start container")
	defer func() {
		terminateErr := container.Terminate(ctx)
		if terminateErr != nil {
			t.Logf("Warning: failed to terminate container: %v", terminateErr)
		}
	}()

	// Create gRPC client
	client, conn, err := createGRPCClient(ctx, container.GRPCAddr)
	require.NoError(t, err, "Failed to create gRPC client")
	defer conn.Close()

	// Wait for proxy to be ready
	err = container.WaitForReady(ctx, 30*time.Second)
	require.NoError(t, err, "Proxy not ready")

	// Verify connection works via Health RPC
	resp, err := client.Health(ctx, &routingv1.HealthRequest{})
	require.NoError(t, err, "Health RPC failed")
	assert.True(t, resp.GetHealthy(), "Proxy should be healthy")
}

func TestGRPC_Health(t *testing.T) {
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

	// Call Health multiple times to verify stability
	for i := range 3 {
		resp, err := client.Health(ctx, &routingv1.HealthRequest{})
		require.NoError(t, err, "Health call %d failed", i+1)
		assert.True(t, resp.GetHealthy(), "Health call %d: proxy should be healthy", i+1)
	}
}
