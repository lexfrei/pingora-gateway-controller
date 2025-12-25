// Package metrics provides Prometheus metrics instrumentation for the controller.
package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector provides metrics recording interface.
// This allows components to record metrics without direct prometheus dependency.
type Collector interface {
	// Sync metrics
	RecordSyncDuration(ctx context.Context, status string, duration time.Duration)
	RecordSyncedRoutes(ctx context.Context, routeType string, count int)
	RecordIngressRules(ctx context.Context, count int)
	RecordFailedBackendRefs(ctx context.Context, routeType string, count int)
	RecordSyncError(ctx context.Context, errorType string)

	// Ingress builder metrics
	RecordIngressBuildDuration(ctx context.Context, routeType string, duration time.Duration)
	RecordBackendRefValidation(ctx context.Context, routeType, result, reason string)

	// gRPC metrics (Pingora proxy communication)
	RecordGRPCCall(ctx context.Context, method, status string, duration time.Duration)
	RecordGRPCError(ctx context.Context, method, errorType string)
}

// prometheusCollector implements Collector using Prometheus metrics.
type prometheusCollector struct {
	// Sync metrics
	syncDuration      *prometheus.HistogramVec
	syncedRoutes      *prometheus.GaugeVec
	ingressRulesTotal prometheus.Gauge
	failedBackendRefs *prometheus.GaugeVec
	syncErrorsTotal   *prometheus.CounterVec

	// Ingress builder metrics
	ingressBuildDuration *prometheus.HistogramVec
	backendRefValidation *prometheus.CounterVec

	// gRPC metrics
	grpcDuration    *prometheus.HistogramVec
	grpcCallsTotal  *prometheus.CounterVec
	grpcErrorsTotal *prometheus.CounterVec
}

// NewCollector creates a new Prometheus metrics collector and registers metrics.
func NewCollector(reg prometheus.Registerer) Collector {
	c := &prometheusCollector{}
	c.initSyncMetrics()
	c.initIngressMetrics()
	c.initGRPCMetrics()
	c.register(reg)

	return c
}

// RecordSyncDuration records the duration of a sync operation.
func (c *prometheusCollector) RecordSyncDuration(_ context.Context, status string, duration time.Duration) {
	c.syncDuration.WithLabelValues(status).Observe(duration.Seconds())
}

// RecordSyncedRoutes records the number of synced routes by type.
func (c *prometheusCollector) RecordSyncedRoutes(_ context.Context, routeType string, count int) {
	c.syncedRoutes.WithLabelValues(routeType).Set(float64(count))
}

// RecordIngressRules records the total number of ingress rules.
func (c *prometheusCollector) RecordIngressRules(_ context.Context, count int) {
	c.ingressRulesTotal.Set(float64(count))
}

// RecordFailedBackendRefs records the number of failed backend references.
func (c *prometheusCollector) RecordFailedBackendRefs(_ context.Context, routeType string, count int) {
	c.failedBackendRefs.WithLabelValues(routeType).Set(float64(count))
}

// RecordSyncError records a sync error by type.
func (c *prometheusCollector) RecordSyncError(_ context.Context, errorType string) {
	c.syncErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordIngressBuildDuration records the duration of ingress rule building.
func (c *prometheusCollector) RecordIngressBuildDuration(
	_ context.Context,
	routeType string,
	duration time.Duration,
) {
	c.ingressBuildDuration.WithLabelValues(routeType).Observe(duration.Seconds())
}

// RecordBackendRefValidation records a backend reference validation result.
func (c *prometheusCollector) RecordBackendRefValidation(_ context.Context, routeType, result, reason string) {
	c.backendRefValidation.WithLabelValues(routeType, result, reason).Inc()
}

// RecordGRPCCall records a gRPC call to the Pingora proxy.
func (c *prometheusCollector) RecordGRPCCall(
	_ context.Context,
	method, status string,
	duration time.Duration,
) {
	c.grpcDuration.WithLabelValues(method).Observe(duration.Seconds())
	c.grpcCallsTotal.WithLabelValues(method, status).Inc()
}

// RecordGRPCError records a gRPC error.
func (c *prometheusCollector) RecordGRPCError(_ context.Context, method, errorType string) {
	c.grpcErrorsTotal.WithLabelValues(method, errorType).Inc()
}

func (c *prometheusCollector) initSyncMetrics() {
	c.syncDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pingora_sync_duration_seconds",
			Help:    "Duration of route synchronization to Pingora proxy",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"status"},
	)
	c.syncedRoutes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pingora_synced_routes",
			Help: "Number of routes synced by type",
		},
		[]string{"type"},
	)
	c.ingressRulesTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "pingora_ingress_rules",
			Help: "Total ingress rules in proxy config",
		},
	)
	c.failedBackendRefs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pingora_failed_backend_refs",
			Help: "Number of failed backend references",
		},
		[]string{"type"},
	)
	c.syncErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pingora_sync_errors_total",
			Help: "Total sync errors by type",
		},
		[]string{"error_type"},
	)
}

func (c *prometheusCollector) initIngressMetrics() {
	c.ingressBuildDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pingora_ingress_build_duration_seconds",
			Help:    "Duration of ingress rule building",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		},
		[]string{"type"},
	)
	c.backendRefValidation = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pingora_backend_ref_validation_total",
			Help: "Backend reference validation results",
		},
		[]string{"type", "result", "reason"},
	)
}

func (c *prometheusCollector) initGRPCMetrics() {
	c.grpcDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pingora_grpc_duration_seconds",
			Help:    "Duration of gRPC calls to Pingora proxy",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method"},
	)
	c.grpcCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pingora_grpc_calls_total",
			Help: "Total gRPC calls to Pingora proxy",
		},
		[]string{"method", "status"},
	)
	c.grpcErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pingora_grpc_errors_total",
			Help: "Total gRPC errors by type",
		},
		[]string{"method", "error_type"},
	)
}

func (c *prometheusCollector) register(reg prometheus.Registerer) {
	reg.MustRegister(
		c.syncDuration,
		c.syncedRoutes,
		c.ingressRulesTotal,
		c.failedBackendRefs,
		c.syncErrorsTotal,
		c.ingressBuildDuration,
		c.backendRefValidation,
		c.grpcDuration,
		c.grpcCallsTotal,
		c.grpcErrorsTotal,
	)
}

// NoopCollector is a no-op implementation of Collector for testing.
type NoopCollector struct{}

// NewNoopCollector creates a new no-op collector.
func NewNoopCollector() *NoopCollector {
	return &NoopCollector{}
}

// RecordSyncDuration is a no-op.
func (c *NoopCollector) RecordSyncDuration(_ context.Context, _ string, _ time.Duration) {}

// RecordSyncedRoutes is a no-op.
func (c *NoopCollector) RecordSyncedRoutes(_ context.Context, _ string, _ int) {}

// RecordIngressRules is a no-op.
func (c *NoopCollector) RecordIngressRules(_ context.Context, _ int) {}

// RecordFailedBackendRefs is a no-op.
func (c *NoopCollector) RecordFailedBackendRefs(_ context.Context, _ string, _ int) {}

// RecordSyncError is a no-op.
func (c *NoopCollector) RecordSyncError(_ context.Context, _ string) {}

// RecordIngressBuildDuration is a no-op.
func (c *NoopCollector) RecordIngressBuildDuration(_ context.Context, _ string, _ time.Duration) {}

// RecordBackendRefValidation is a no-op.
func (c *NoopCollector) RecordBackendRefValidation(_ context.Context, _, _, _ string) {}

// RecordGRPCCall is a no-op.
func (c *NoopCollector) RecordGRPCCall(_ context.Context, _, _ string, _ time.Duration) {}

// RecordGRPCError is a no-op.
func (c *NoopCollector) RecordGRPCError(_ context.Context, _, _ string) {}
