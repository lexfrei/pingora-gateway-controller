//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	grpcPort   = "50051/tcp"
	httpPort   = "8080/tcp"
	healthPort = "9090/tcp"
)

// PingoraContainer represents a running Pingora proxy container.
type PingoraContainer struct {
	container  testcontainers.Container
	GRPCAddr   string // e.g., "localhost:55001"
	HTTPAddr   string // e.g., "localhost:8081"
	HealthAddr string // e.g., "localhost:9091"
}

// StartPingoraContainer starts a new Pingora proxy container.
func StartPingoraContainer(ctx context.Context) (*PingoraContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        pingoraImage,
		ExposedPorts: []string{grpcPort, httpPort, healthPort},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/healthz").WithPort(healthPort).WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get mapped ports
	grpcMapped, err := container.MappedPort(ctx, grpcPort)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get gRPC port: %w", err)
	}

	httpMapped, err := container.MappedPort(ctx, httpPort)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get HTTP port: %w", err)
	}

	healthMapped, err := container.MappedPort(ctx, healthPort)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get health port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	return &PingoraContainer{
		container:  container,
		GRPCAddr:   fmt.Sprintf("%s:%s", host, grpcMapped.Port()),
		HTTPAddr:   fmt.Sprintf("%s:%s", host, httpMapped.Port()),
		HealthAddr: fmt.Sprintf("%s:%s", host, healthMapped.Port()),
	}, nil
}

// Terminate stops and removes the container.
func (p *PingoraContainer) Terminate(ctx context.Context) error {
	if p.container == nil {
		return nil
	}

	err := p.container.Terminate(ctx)
	if err != nil {
		return fmt.Errorf("failed to terminate container: %w", err)
	}

	return nil
}

// ErrProxyNotReady is returned when the proxy fails to become ready within the timeout.
var ErrProxyNotReady = fmt.Errorf("proxy not ready")

// WaitForReady waits for the proxy to be ready by polling the health endpoint.
func (p *PingoraContainer) WaitForReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+p.HealthAddr+"/healthz", nil)
		if err != nil {
			return fmt.Errorf("failed to create health request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for proxy: %w", ctx.Err())
		case <-time.After(500 * time.Millisecond):
			// retry
		}
	}

	return fmt.Errorf("%w after %v", ErrProxyNotReady, timeout)
}

// Logs returns the container logs.
func (p *PingoraContainer) Logs(ctx context.Context) (string, error) {
	reader, err := p.container.Logs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	buf := make([]byte, 64*1024)
	n, _ := reader.Read(buf)

	return string(buf[:n]), nil
}
