package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorInterface(t *testing.T) {
	t.Parallel()

	// Verify that prometheusCollector implements Collector interface
	var _ Collector = (*prometheusCollector)(nil)
	var _ Collector = (*NoopCollector)(nil)
}

func TestNewCollector(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg)

	require.NotNil(t, collector)
	assert.IsType(t, &prometheusCollector{}, collector)
}

func TestNoopCollector(t *testing.T) {
	t.Parallel()

	collector := NewNoopCollector()
	require.NotNil(t, collector)

	ctx := context.Background()

	// All methods should not panic
	assert.NotPanics(t, func() {
		collector.RecordSyncDuration(ctx, "success", time.Second)
		collector.RecordSyncedRoutes(ctx, "http", 5)
		collector.RecordIngressRules(ctx, 10)
		collector.RecordFailedBackendRefs(ctx, "http", 2)
		collector.RecordSyncError(ctx, "timeout")
		collector.RecordIngressBuildDuration(ctx, "http", time.Millisecond*100)
		collector.RecordBackendRefValidation(ctx, "http", "accepted", "")
		collector.RecordGRPCCall(ctx, "UpdateRoutes", "success", time.Second)
		collector.RecordGRPCError(ctx, "UpdateRoutes", "timeout")
	})
}

func TestMetricsRegistration(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	// Trigger all metrics to be collected at least once
	collector.RecordSyncDuration(ctx, "success", time.Second)
	collector.RecordSyncedRoutes(ctx, "http", 1)
	collector.RecordIngressRules(ctx, 1)
	collector.RecordFailedBackendRefs(ctx, "http", 0)
	collector.RecordSyncError(ctx, "test")
	collector.RecordIngressBuildDuration(ctx, "http", time.Millisecond)
	collector.RecordBackendRefValidation(ctx, "http", "accepted", "")
	collector.RecordGRPCCall(ctx, "UpdateRoutes", "success", time.Second)
	collector.RecordGRPCError(ctx, "UpdateRoutes", "test")

	// Verify metrics are registered
	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	expectedMetrics := []string{
		// Sync metrics
		"pingora_sync_duration_seconds",
		"pingora_synced_routes",
		"pingora_ingress_rules",
		"pingora_failed_backend_refs",
		"pingora_sync_errors_total",
		// Ingress builder metrics
		"pingora_ingress_build_duration_seconds",
		"pingora_backend_ref_validation_total",
		// gRPC metrics
		"pingora_grpc_duration_seconds",
		"pingora_grpc_calls_total",
		"pingora_grpc_errors_total",
	}

	registeredMetrics := make(map[string]bool)
	for _, mf := range metricFamilies {
		registeredMetrics[mf.GetName()] = true
	}

	for _, expected := range expectedMetrics {
		assert.True(t, registeredMetrics[expected], "metric %s should be registered", expected)
	}
}

func TestRecordSyncDuration(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordSyncDuration(ctx, "success", time.Second)

	// Check that histogram was observed
	count := testutil.CollectAndCount(collector.syncDuration)
	assert.Equal(t, 1, count)
}

func TestRecordSyncedRoutes(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordSyncedRoutes(ctx, "http", 5)
	collector.RecordSyncedRoutes(ctx, "grpc", 3)

	httpCount := testutil.ToFloat64(collector.syncedRoutes.WithLabelValues("http"))
	grpcCount := testutil.ToFloat64(collector.syncedRoutes.WithLabelValues("grpc"))

	assert.Equal(t, float64(5), httpCount)
	assert.Equal(t, float64(3), grpcCount)
}

func TestRecordIngressRules(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordIngressRules(ctx, 10)

	count := testutil.ToFloat64(collector.ingressRulesTotal)
	assert.Equal(t, float64(10), count)
}

func TestRecordFailedBackendRefs(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordFailedBackendRefs(ctx, "http", 2)

	count := testutil.ToFloat64(collector.failedBackendRefs.WithLabelValues("http"))
	assert.Equal(t, float64(2), count)
}

func TestRecordSyncError(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordSyncError(ctx, "timeout")
	collector.RecordSyncError(ctx, "timeout")
	collector.RecordSyncError(ctx, "network")

	timeoutCount := testutil.ToFloat64(collector.syncErrorsTotal.WithLabelValues("timeout"))
	networkCount := testutil.ToFloat64(collector.syncErrorsTotal.WithLabelValues("network"))

	assert.Equal(t, float64(2), timeoutCount)
	assert.Equal(t, float64(1), networkCount)
}

func TestRecordIngressBuildDuration(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordIngressBuildDuration(ctx, "http", time.Millisecond*100)

	// Check histogram was observed
	count := testutil.CollectAndCount(collector.ingressBuildDuration)
	assert.Equal(t, 1, count)
}

func TestRecordBackendRefValidation(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordBackendRefValidation(ctx, "http", "accepted", "")
	collector.RecordBackendRefValidation(ctx, "http", "rejected", "not_found")

	acceptedCount := testutil.ToFloat64(collector.backendRefValidation.WithLabelValues("http", "accepted", ""))
	rejectedCount := testutil.ToFloat64(collector.backendRefValidation.WithLabelValues("http", "rejected", "not_found"))

	assert.Equal(t, float64(1), acceptedCount)
	assert.Equal(t, float64(1), rejectedCount)
}

func TestRecordGRPCCall(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordGRPCCall(ctx, "UpdateRoutes", "success", time.Second)

	// Check histogram and counter
	durationCount := testutil.CollectAndCount(collector.grpcDuration)
	callsCount := testutil.ToFloat64(collector.grpcCallsTotal.WithLabelValues("UpdateRoutes", "success"))

	assert.Equal(t, 1, durationCount)
	assert.Equal(t, float64(1), callsCount)
}

func TestRecordGRPCError(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	collector.RecordGRPCError(ctx, "UpdateRoutes", "connection_refused")

	count := testutil.ToFloat64(collector.grpcErrorsTotal.WithLabelValues("UpdateRoutes", "connection_refused"))
	assert.Equal(t, float64(1), count)
}

func TestHistogramBuckets(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	collector := NewCollector(reg).(*prometheusCollector)
	ctx := context.Background()

	// Record sync duration of 1 second
	collector.RecordSyncDuration(ctx, "success", time.Second)

	// Verify histogram was collected (bucket verification via lint)
	count := testutil.CollectAndCount(collector.syncDuration)
	assert.Equal(t, 1, count)
}
