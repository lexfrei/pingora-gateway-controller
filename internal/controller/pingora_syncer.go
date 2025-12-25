package controller

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lexfrei/pingora-gateway-controller/internal/config"
	"github.com/lexfrei/pingora-gateway-controller/internal/metrics"
	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

// PingoraSyncer handles synchronization of routes to Pingora proxy via gRPC.
type PingoraSyncer struct {
	resolver *config.PingoraResolver
	metrics  metrics.Collector

	// Connection state
	mu         sync.RWMutex
	conn       *grpc.ClientConn
	grpcClient routingv1.RoutingServiceClient
	configName string

	// Version tracking for optimistic concurrency
	version atomic.Uint64
}

// NewPingoraSyncer creates a new PingoraSyncer.
func NewPingoraSyncer(
	k8sClient client.Client,
	defaultNamespace string,
	metricsCollector metrics.Collector,
) *PingoraSyncer {
	return &PingoraSyncer{
		resolver: config.NewPingoraResolver(k8sClient, defaultNamespace),
		metrics:  metricsCollector,
	}
}

// Connect establishes a gRPC connection to the Pingora proxy.
func (s *PingoraSyncer) Connect(ctx context.Context, gatewayClassName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close existing connection if any
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			log.FromContext(ctx).Error(err, "failed to close existing connection")
		}
	}

	// Resolve config
	resolved, err := s.resolver.ResolveFromGatewayClassName(ctx, gatewayClassName)
	if err != nil {
		return errors.Wrap(err, "failed to resolve Pingora config")
	}

	// Create new connection
	conn, err := s.resolver.CreateGRPCConnection(ctx, resolved)
	if err != nil {
		return errors.Wrap(err, "failed to create gRPC connection")
	}

	s.conn = conn
	s.grpcClient = s.resolver.CreateRoutingClient(conn)
	s.configName = resolved.ConfigName

	return nil
}

// Close closes the gRPC connection.
func (s *PingoraSyncer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.grpcClient = nil

		return err //nolint:wrapcheck // simple close error
	}

	return nil
}

// IsConnected returns whether a connection is established.
func (s *PingoraSyncer) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.grpcClient != nil
}

// SyncRoutes sends all routes to the Pingora proxy.
func (s *PingoraSyncer) SyncRoutes(
	ctx context.Context,
	httpRoutes []*routingv1.HTTPRoute,
	grpcRoutes []*routingv1.GRPCRoute,
) error {
	s.mu.RLock()
	rpcClient := s.grpcClient
	s.mu.RUnlock()

	if rpcClient == nil {
		return errors.New("not connected to Pingora proxy")
	}

	// Increment version
	version := s.version.Add(1)

	req := &routingv1.UpdateRoutesRequest{
		HttpRoutes: httpRoutes,
		GrpcRoutes: grpcRoutes,
		Version:    version,
	}

	startTime := time.Now()
	resp, err := rpcClient.UpdateRoutes(ctx, req)
	duration := time.Since(startTime)

	if err != nil {
		s.metrics.RecordGRPCCall(ctx, "UpdateRoutes", "error", duration)

		return errors.Wrap(err, "failed to update routes")
	}

	if !resp.GetSuccess() {
		s.metrics.RecordGRPCCall(ctx, "UpdateRoutes", "failed", duration)

		return errors.Newf("route update failed: %s", resp.GetError()) //nolint:wrapcheck // Newf creates new error
	}

	s.metrics.RecordGRPCCall(ctx, "UpdateRoutes", "success", duration)
	log.FromContext(ctx).Info("routes synced successfully",
		"httpRouteCount", resp.GetHttpRouteCount(),
		"grpcRouteCount", resp.GetGrpcRouteCount(),
		"version", resp.GetAppliedVersion(),
	)

	return nil
}

// GetRoutes retrieves the current routes from the Pingora proxy.
//
//nolint:dupl // similar pattern to Health() is intentional
func (s *PingoraSyncer) GetRoutes(ctx context.Context) (*routingv1.GetRoutesResponse, error) {
	s.mu.RLock()
	rpcClient := s.grpcClient
	s.mu.RUnlock()

	if rpcClient == nil {
		return nil, errors.New("not connected to Pingora proxy")
	}

	startTime := time.Now()
	resp, err := rpcClient.GetRoutes(ctx, &routingv1.GetRoutesRequest{})
	duration := time.Since(startTime)

	if err != nil {
		s.metrics.RecordGRPCCall(ctx, "GetRoutes", "error", duration)

		return nil, errors.Wrap(err, "failed to get routes")
	}

	s.metrics.RecordGRPCCall(ctx, "GetRoutes", "success", duration)

	return resp, nil
}

// Health checks the health of the Pingora proxy.
//
//nolint:dupl // similar pattern to GetRoutes() is intentional
func (s *PingoraSyncer) Health(ctx context.Context) (*routingv1.HealthResponse, error) {
	s.mu.RLock()
	rpcClient := s.grpcClient
	s.mu.RUnlock()

	if rpcClient == nil {
		return nil, errors.New("not connected to Pingora proxy")
	}

	startTime := time.Now()
	resp, err := rpcClient.Health(ctx, &routingv1.HealthRequest{})
	duration := time.Since(startTime)

	if err != nil {
		s.metrics.RecordGRPCCall(ctx, "Health", "error", duration)

		return nil, errors.Wrap(err, "failed to check health")
	}

	s.metrics.RecordGRPCCall(ctx, "Health", "success", duration)

	return resp, nil
}

// GetVersion returns the current version counter.
func (s *PingoraSyncer) GetVersion() uint64 {
	return s.version.Load()
}

// GetConfigName returns the name of the current PingoraConfig.
func (s *PingoraSyncer) GetConfigName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.configName
}
