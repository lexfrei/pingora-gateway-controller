package controller

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/lexfrei/pingora-gateway-controller/internal/config"
	pingoraingress "github.com/lexfrei/pingora-gateway-controller/internal/ingress"
	"github.com/lexfrei/pingora-gateway-controller/internal/logging"
	"github.com/lexfrei/pingora-gateway-controller/internal/metrics"
	"github.com/lexfrei/pingora-gateway-controller/internal/routebinding"
	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

const (
	// apiErrorRequeueDelay is the delay before retrying when API errors occur.
	apiErrorRequeueDelay = 30 * time.Second
)

// SyncResult holds the results of a route synchronization.
type SyncResult struct {
	HTTPRoutes        []gatewayv1.HTTPRoute
	GRPCRoutes        []gatewayv1.GRPCRoute
	HTTPRouteBindings map[string]routeBindingInfo
	GRPCRouteBindings map[string]routeBindingInfo
}

// routeBindingInfo holds binding validation results for a route.
type routeBindingInfo struct {
	bindingResults map[int]routebinding.BindingResult
}

// PingoraRouteSyncer provides unified synchronization of HTTPRoute and GRPCRoute
// resources to Pingora proxy via gRPC.
//
// Both HTTPRouteReconciler and GRPCRouteReconciler use this to sync routes,
// ensuring that all route types are collected and synchronized together.
type PingoraRouteSyncer struct {
	client.Client

	Scheme           *runtime.Scheme
	ClusterDomain    string
	GatewayClassName string
	ConfigResolver   *config.PingoraResolver
	Metrics          metrics.Collector
	Logger           *slog.Logger

	builder          *pingoraingress.PingoraBuilder
	bindingValidator *routebinding.Validator

	// gRPC connection state
	connMu     sync.RWMutex
	conn       *grpc.ClientConn
	grpcClient routingv1.RoutingServiceClient
	configName string

	// Version tracking for optimistic concurrency
	version atomic.Uint64

	// syncMu protects concurrent calls to SyncAllRoutes.
	// Both HTTPRouteReconciler and GRPCRouteReconciler may call SyncAllRoutes
	// concurrently, and this mutex ensures serialized access to gRPC calls.
	syncMu sync.Mutex
}

// NewPingoraRouteSyncer creates a new PingoraRouteSyncer.
func NewPingoraRouteSyncer(
	c client.Client,
	scheme *runtime.Scheme,
	clusterDomain string,
	gatewayClassName string,
	configResolver *config.PingoraResolver,
	metricsCollector metrics.Collector,
	logger *slog.Logger,
) *PingoraRouteSyncer {
	if logger == nil {
		logger = slog.Default()
	}

	componentLogger := logger.With("component", "pingora-route-syncer")

	return &PingoraRouteSyncer{
		Client:           c,
		Scheme:           scheme,
		ClusterDomain:    clusterDomain,
		GatewayClassName: gatewayClassName,
		ConfigResolver:   configResolver,
		Metrics:          metricsCollector,
		Logger:           componentLogger,
		builder:          pingoraingress.NewPingoraBuilder(clusterDomain),
		bindingValidator: routebinding.NewValidator(c),
	}
}

// Connect establishes a gRPC connection to the Pingora proxy.
func (s *PingoraRouteSyncer) Connect(ctx context.Context) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	// Close existing connection if any
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			s.Logger.Error("failed to close existing connection", "error", err)
		}
	}

	// Resolve config
	resolved, err := s.ConfigResolver.ResolveFromGatewayClassName(ctx, s.GatewayClassName)
	if err != nil {
		return errors.Wrap(err, "failed to resolve Pingora config")
	}

	// Create new connection
	conn, err := s.ConfigResolver.CreateGRPCConnection(ctx, resolved)
	if err != nil {
		return errors.Wrap(err, "failed to create gRPC connection")
	}

	s.conn = conn
	s.grpcClient = s.ConfigResolver.CreateRoutingClient(conn)
	s.configName = resolved.ConfigName

	s.Logger.Info("connected to Pingora proxy", "address", resolved.Address)

	return nil
}

// Close closes the gRPC connection.
func (s *PingoraRouteSyncer) Close() error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.grpcClient = nil

		return err //nolint:wrapcheck // simple close error
	}

	return nil
}

// IsConnected returns whether a connection is established.
func (s *PingoraRouteSyncer) IsConnected() bool {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	return s.grpcClient != nil
}

// SyncAllRoutes synchronizes all HTTPRoute and GRPCRoute resources to Pingora proxy.
//
//nolint:funlen // complex sync logic requires length
func (s *PingoraRouteSyncer) SyncAllRoutes(ctx context.Context) (ctrl.Result, *SyncResult, error) {
	// Serialize concurrent sync calls to prevent race conditions when
	// both HTTPRouteReconciler and GRPCRouteReconciler trigger syncs.
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	startTime := time.Now()

	// Prefer context logger (with reconcile ID) over struct logger
	logger := logging.FromContext(ctx)
	if logger == slog.Default() {
		logger = s.Logger
	}

	// Ensure we're connected
	if !s.IsConnected() {
		if err := s.Connect(ctx); err != nil {
			logger.Error("failed to connect to Pingora proxy", "error", err)
			s.Metrics.RecordSyncDuration(ctx, "error", time.Since(startTime))
			s.Metrics.RecordSyncError(ctx, "connection_failed")

			return ctrl.Result{RequeueAfter: apiErrorRequeueDelay}, nil, nil
		}
	}

	// Collect all relevant HTTPRoutes with binding validation
	httpRoutes, httpBindings, err := s.getRelevantHTTPRoutes(ctx)
	if err != nil {
		return ctrl.Result{}, nil, errors.Wrap(err, "failed to list httproutes")
	}

	// Collect all relevant GRPCRoutes with binding validation
	grpcRoutes, grpcBindings, err := s.getRelevantGRPCRoutes(ctx)
	if err != nil {
		return ctrl.Result{}, nil, errors.Wrap(err, "failed to list grpcroutes")
	}

	logger.Info("syncing routes to Pingora",
		"httpRoutes", len(httpRoutes),
		"grpcRoutes", len(grpcRoutes),
	)

	// Build Pingora route configurations
	pingoraHTTPRoutes := make([]*routingv1.HTTPRoute, 0, len(httpRoutes))
	for i := range httpRoutes {
		pingoraHTTPRoutes = append(pingoraHTTPRoutes, s.builder.BuildHTTPRoute(&httpRoutes[i]))
	}

	pingoraGRPCRoutes := make([]*routingv1.GRPCRoute, 0, len(grpcRoutes))
	for i := range grpcRoutes {
		pingoraGRPCRoutes = append(pingoraGRPCRoutes, s.builder.BuildGRPCRoute(&grpcRoutes[i]))
	}

	// Send routes to Pingora via gRPC
	version := s.version.Add(1)

	req := &routingv1.UpdateRoutesRequest{
		HttpRoutes: pingoraHTTPRoutes,
		GrpcRoutes: pingoraGRPCRoutes,
		Version:    version,
	}

	s.connMu.RLock()
	grpcClient := s.grpcClient
	s.connMu.RUnlock()

	if grpcClient == nil {
		logger.Error("gRPC client is nil")
		s.Metrics.RecordSyncDuration(ctx, "error", time.Since(startTime))
		s.Metrics.RecordSyncError(ctx, "not_connected")

		return ctrl.Result{RequeueAfter: apiErrorRequeueDelay}, nil, nil
	}

	grpcStart := time.Now()
	resp, err := grpcClient.UpdateRoutes(ctx, req)
	grpcDuration := time.Since(grpcStart)

	if err != nil {
		s.Metrics.RecordGRPCCall(ctx, "UpdateRoutes", "error", grpcDuration)
		s.Metrics.RecordSyncDuration(ctx, "error", time.Since(startTime))
		s.Metrics.RecordSyncError(ctx, "grpc_error")
		logger.Error("failed to update routes via gRPC", "error", err)

		// Try to reconnect on next sync
		s.connMu.Lock()

		if s.conn != nil {
			_ = s.conn.Close()
			s.conn = nil
			s.grpcClient = nil
		}

		s.connMu.Unlock()

		result := &SyncResult{
			HTTPRoutes:        httpRoutes,
			GRPCRoutes:        grpcRoutes,
			HTTPRouteBindings: httpBindings,
			GRPCRouteBindings: grpcBindings,
		}

		return ctrl.Result{RequeueAfter: apiErrorRequeueDelay}, result, errors.Wrap(err, "failed to update routes via gRPC")
	}

	if !resp.GetSuccess() {
		s.Metrics.RecordGRPCCall(ctx, "UpdateRoutes", "failed", grpcDuration)
		s.Metrics.RecordSyncDuration(ctx, "error", time.Since(startTime))
		s.Metrics.RecordSyncError(ctx, "update_failed")
		logger.Error("route update failed", "error", resp.GetError())

		result := &SyncResult{
			HTTPRoutes:        httpRoutes,
			GRPCRoutes:        grpcRoutes,
			HTTPRouteBindings: httpBindings,
			GRPCRouteBindings: grpcBindings,
		}

		//nolint:wrapcheck // Newf creates new error, not wrapping
		return ctrl.Result{RequeueAfter: apiErrorRequeueDelay}, result, errors.Newf("route update failed: %s", resp.GetError())
	}

	s.Metrics.RecordGRPCCall(ctx, "UpdateRoutes", "success", grpcDuration)
	logger.Info("successfully updated routes in Pingora",
		"httpRouteCount", resp.GetHttpRouteCount(),
		"grpcRouteCount", resp.GetGrpcRouteCount(),
		"version", resp.GetAppliedVersion(),
	)

	// Record success metrics
	s.Metrics.RecordSyncDuration(ctx, "success", time.Since(startTime))
	s.Metrics.RecordSyncedRoutes(ctx, "http", len(httpRoutes))
	s.Metrics.RecordSyncedRoutes(ctx, "grpc", len(grpcRoutes))

	result := &SyncResult{
		HTTPRoutes:        httpRoutes,
		GRPCRoutes:        grpcRoutes,
		HTTPRouteBindings: httpBindings,
		GRPCRouteBindings: grpcBindings,
	}

	return ctrl.Result{}, result, nil
}

//nolint:funlen,dupl // complex binding validation logic; similar to GRPC but for HTTP types
func (s *PingoraRouteSyncer) getRelevantHTTPRoutes(
	ctx context.Context,
) ([]gatewayv1.HTTPRoute, map[string]routeBindingInfo, error) {
	// Prefer context logger (with reconcile ID) over struct logger
	logger := logging.FromContext(ctx)
	if logger == slog.Default() {
		logger = s.Logger
	}

	var routeList gatewayv1.HTTPRouteList

	err := s.List(ctx, &routeList)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list httproutes")
	}

	var relevantRoutes []gatewayv1.HTTPRoute

	bindings := make(map[string]routeBindingInfo)

	for i := range routeList.Items {
		route := &routeList.Items[i]
		routeKey := route.Namespace + "/" + route.Name
		bindingInfo := routeBindingInfo{
			bindingResults: make(map[int]routebinding.BindingResult),
		}

		hasAcceptedBinding := false

		for refIdx, ref := range route.Spec.ParentRefs {
			if ref.Kind != nil && *ref.Kind != kindGateway {
				continue
			}

			namespace := route.Namespace
			if ref.Namespace != nil {
				namespace = string(*ref.Namespace)
			}

			var gateway gatewayv1.Gateway

			getErr := s.Get(ctx, client.ObjectKey{Name: string(ref.Name), Namespace: namespace}, &gateway)
			if getErr != nil {
				continue
			}

			if gateway.Spec.GatewayClassName != gatewayv1.ObjectName(s.GatewayClassName) {
				continue
			}

			routeInfo := &routebinding.RouteInfo{
				Name:        route.Name,
				Namespace:   route.Namespace,
				Hostnames:   route.Spec.Hostnames,
				Kind:        routebinding.KindHTTPRoute,
				SectionName: ref.SectionName,
			}

			result, bindErr := s.bindingValidator.ValidateBinding(ctx, &gateway, routeInfo)
			if bindErr != nil {
				logger.Error("failed to validate route binding",
					"route", routeKey,
					"gateway", gateway.Name,
					"error", bindErr)

				continue
			}

			bindingInfo.bindingResults[refIdx] = result

			if result.Accepted {
				hasAcceptedBinding = true
			}
		}

		bindings[routeKey] = bindingInfo

		if hasAcceptedBinding {
			relevantRoutes = append(relevantRoutes, routeList.Items[i])
		}
	}

	return relevantRoutes, bindings, nil
}

//nolint:funlen,dupl // complex binding validation logic; similar to HTTP but for GRPC types
func (s *PingoraRouteSyncer) getRelevantGRPCRoutes(
	ctx context.Context,
) ([]gatewayv1.GRPCRoute, map[string]routeBindingInfo, error) {
	// Prefer context logger (with reconcile ID) over struct logger
	logger := logging.FromContext(ctx)
	if logger == slog.Default() {
		logger = s.Logger
	}

	var routeList gatewayv1.GRPCRouteList

	err := s.List(ctx, &routeList)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list grpcroutes")
	}

	var relevantRoutes []gatewayv1.GRPCRoute

	bindings := make(map[string]routeBindingInfo)

	for i := range routeList.Items {
		route := &routeList.Items[i]
		routeKey := route.Namespace + "/" + route.Name
		bindingInfo := routeBindingInfo{
			bindingResults: make(map[int]routebinding.BindingResult),
		}

		hasAcceptedBinding := false

		for refIdx, ref := range route.Spec.ParentRefs {
			if ref.Kind != nil && *ref.Kind != kindGateway {
				continue
			}

			namespace := route.Namespace
			if ref.Namespace != nil {
				namespace = string(*ref.Namespace)
			}

			var gateway gatewayv1.Gateway

			getErr := s.Get(ctx, client.ObjectKey{Name: string(ref.Name), Namespace: namespace}, &gateway)
			if getErr != nil {
				continue
			}

			if gateway.Spec.GatewayClassName != gatewayv1.ObjectName(s.GatewayClassName) {
				continue
			}

			routeInfo := &routebinding.RouteInfo{
				Name:        route.Name,
				Namespace:   route.Namespace,
				Hostnames:   route.Spec.Hostnames,
				Kind:        routebinding.KindGRPCRoute,
				SectionName: ref.SectionName,
			}

			result, bindErr := s.bindingValidator.ValidateBinding(ctx, &gateway, routeInfo)
			if bindErr != nil {
				logger.Error("failed to validate route binding",
					"route", routeKey,
					"gateway", gateway.Name,
					"error", bindErr)

				continue
			}

			bindingInfo.bindingResults[refIdx] = result

			if result.Accepted {
				hasAcceptedBinding = true
			}
		}

		bindings[routeKey] = bindingInfo

		if hasAcceptedBinding {
			relevantRoutes = append(relevantRoutes, routeList.Items[i])
		}
	}

	return relevantRoutes, bindings, nil
}

// GetConfigName returns the name of the current PingoraConfig.
func (s *PingoraRouteSyncer) GetConfigName() string {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	return s.configName
}

// GetVersion returns the current version counter.
func (s *PingoraRouteSyncer) GetVersion() uint64 {
	return s.version.Load()
}
