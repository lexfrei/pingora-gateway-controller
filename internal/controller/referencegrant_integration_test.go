package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/lexfrei/pingora-gateway-controller/api/v1alpha1"
	"github.com/lexfrei/pingora-gateway-controller/internal/referencegrant"
)

// TestReferenceGrant_Validator_Direct tests the validator directly.
func TestReferenceGrant_Validator_Direct(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, gatewayv1.Install(scheme))
	require.NoError(t, gatewayv1beta1.Install(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// ReferenceGrant in "backend" namespace
	refGrant := &gatewayv1beta1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-grant",
			Namespace: "backend",
		},
		Spec: gatewayv1beta1.ReferenceGrantSpec{
			From: []gatewayv1beta1.ReferenceGrantFrom{
				{
					Group:     gatewayv1.GroupName,
					Kind:      "HTTPRoute",
					Namespace: "default",
				},
			},
			To: []gatewayv1beta1.ReferenceGrantTo{
				{
					Group: "",
					Kind:  "Service",
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(refGrant).
		Build()

	validator := referencegrant.NewValidator(fakeClient)

	ctx := context.Background()

	// Test the reference - should be allowed with ReferenceGrant
	allowed, err := validator.IsReferenceAllowed(ctx,
		referencegrant.Reference{
			Group:     gatewayv1.GroupName,
			Kind:      "HTTPRoute",
			Namespace: "default",
			Name:      "test-route",
		},
		referencegrant.Reference{
			Group:     "",
			Kind:      "Service",
			Namespace: "backend",
			Name:      "backend-service",
		},
	)

	require.NoError(t, err)
	assert.True(t, allowed, "Reference should be allowed with valid ReferenceGrant")
}

// TestReferenceGrant_Validator_NoGrant tests that cross-namespace
// references are denied when no ReferenceGrant exists.
func TestReferenceGrant_Validator_NoGrant(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, gatewayv1.Install(scheme))
	require.NoError(t, gatewayv1beta1.Install(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// No ReferenceGrant
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	validator := referencegrant.NewValidator(fakeClient)

	ctx := context.Background()

	// Test the reference - should be denied without ReferenceGrant
	allowed, err := validator.IsReferenceAllowed(ctx,
		referencegrant.Reference{
			Group:     gatewayv1.GroupName,
			Kind:      "HTTPRoute",
			Namespace: "default",
			Name:      "test-route",
		},
		referencegrant.Reference{
			Group:     "",
			Kind:      "Service",
			Namespace: "backend",
			Name:      "backend-service",
		},
	)

	require.NoError(t, err)
	assert.False(t, allowed, "Reference should be denied without ReferenceGrant")
}

// TestReferenceGrant_Validator_SpecificServiceName tests that ReferenceGrant
// can limit access to specific Service names.
func TestReferenceGrant_Validator_SpecificServiceName(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, gatewayv1.Install(scheme))
	require.NoError(t, gatewayv1beta1.Install(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// ReferenceGrant that only allows access to "allowed-service"
	allowedServiceName := gatewayv1.ObjectName("allowed-service")
	refGrant := &gatewayv1beta1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "specific-service-grant",
			Namespace: "backend",
		},
		Spec: gatewayv1beta1.ReferenceGrantSpec{
			From: []gatewayv1beta1.ReferenceGrantFrom{
				{
					Group:     gatewayv1.GroupName,
					Kind:      "HTTPRoute",
					Namespace: "default",
				},
			},
			To: []gatewayv1beta1.ReferenceGrantTo{
				{
					Group: "",
					Kind:  "Service",
					Name:  &allowedServiceName,
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(refGrant).
		Build()

	validator := referencegrant.NewValidator(fakeClient)

	ctx := context.Background()

	// Test reference to allowed-service - should be allowed
	allowedRef, err := validator.IsReferenceAllowed(ctx,
		referencegrant.Reference{
			Group:     gatewayv1.GroupName,
			Kind:      "HTTPRoute",
			Namespace: "default",
			Name:      "test-route",
		},
		referencegrant.Reference{
			Group:     "",
			Kind:      "Service",
			Namespace: "backend",
			Name:      "allowed-service",
		},
	)

	require.NoError(t, err)
	assert.True(t, allowedRef, "Reference to allowed-service should be allowed")

	// Test reference to denied-service - should be denied
	deniedRef, err := validator.IsReferenceAllowed(ctx,
		referencegrant.Reference{
			Group:     gatewayv1.GroupName,
			Kind:      "HTTPRoute",
			Namespace: "default",
			Name:      "test-route",
		},
		referencegrant.Reference{
			Group:     "",
			Kind:      "Service",
			Namespace: "backend",
			Name:      "denied-service",
		},
	)

	require.NoError(t, err)
	assert.False(t, deniedRef, "Reference to denied-service should be denied")
}

// TestReferenceGrant_Validator_GRPCRoute tests ReferenceGrant for GRPCRoute resources.
func TestReferenceGrant_Validator_GRPCRoute(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, gatewayv1.Install(scheme))
	require.NoError(t, gatewayv1beta1.Install(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// ReferenceGrant for GRPCRoute
	refGrant := &gatewayv1beta1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grpc-grant",
			Namespace: "backend",
		},
		Spec: gatewayv1beta1.ReferenceGrantSpec{
			From: []gatewayv1beta1.ReferenceGrantFrom{
				{
					Group:     gatewayv1.GroupName,
					Kind:      "GRPCRoute",
					Namespace: "default",
				},
			},
			To: []gatewayv1beta1.ReferenceGrantTo{
				{
					Group: "",
					Kind:  "Service",
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(refGrant).
		Build()

	validator := referencegrant.NewValidator(fakeClient)

	ctx := context.Background()

	// Test GRPCRoute reference - should be allowed
	allowed, err := validator.IsReferenceAllowed(ctx,
		referencegrant.Reference{
			Group:     gatewayv1.GroupName,
			Kind:      "GRPCRoute",
			Namespace: "default",
			Name:      "test-grpc-route",
		},
		referencegrant.Reference{
			Group:     "",
			Kind:      "Service",
			Namespace: "backend",
			Name:      "grpc-service",
		},
	)

	require.NoError(t, err)
	assert.True(t, allowed, "GRPCRoute reference should be allowed with valid ReferenceGrant")

	// Test HTTPRoute reference with GRPCRoute grant - should be denied
	httpRef, err := validator.IsReferenceAllowed(ctx,
		referencegrant.Reference{
			Group:     gatewayv1.GroupName,
			Kind:      "HTTPRoute",
			Namespace: "default",
			Name:      "test-http-route",
		},
		referencegrant.Reference{
			Group:     "",
			Kind:      "Service",
			Namespace: "backend",
			Name:      "grpc-service",
		},
	)

	require.NoError(t, err)
	assert.False(t, httpRef, "HTTPRoute reference should be denied when only GRPCRoute is granted")
}
