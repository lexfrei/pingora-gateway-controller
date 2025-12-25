package controller

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/lexfrei/pingora-gateway-controller/api/v1alpha1"
	"github.com/lexfrei/pingora-gateway-controller/internal/config"
	"github.com/lexfrei/pingora-gateway-controller/internal/logging"
	"github.com/lexfrei/pingora-gateway-controller/internal/routebinding"
)

const (
	// configErrorRequeueDelay is the delay before retrying when config resolution fails.
	configErrorRequeueDelay = 30 * time.Second
)

// PingoraGatewayReconciler reconciles Gateway resources for the Pingora GatewayClass.
//
// It performs the following functions:
//   - Watches Gateway resources matching the configured GatewayClassName
//   - Reads configuration from PingoraConfig via parametersRef
//   - Updates Gateway status with Pingora proxy connection status
//   - Handles Gateway deletion with proper cleanup
type PingoraGatewayReconciler struct {
	client.Client

	// Scheme is the runtime scheme for API type registration.
	Scheme *runtime.Scheme

	// GatewayClassName is the name of the GatewayClass to watch.
	GatewayClassName string

	// ControllerName is reported in Gateway status conditions.
	ControllerName string

	// ConfigResolver resolves configuration from PingoraConfig.
	ConfigResolver *config.PingoraResolver
}

func (r *PingoraGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var gateway gatewayv1.Gateway

	if err := r.Get(ctx, req.NamespacedName, &gateway); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrap(err, "failed to get gateway")
	}

	if gateway.Spec.GatewayClassName != gatewayv1.ObjectName(r.GatewayClassName) {
		return ctrl.Result{}, nil
	}

	logger.Info("reconciling gateway", "name", gateway.Name, "namespace", gateway.Namespace)

	// Resolve configuration from PingoraConfig
	resolvedConfig, err := r.ConfigResolver.ResolveFromGatewayClassName(ctx, r.GatewayClassName)
	if err != nil {
		logger.Error(err, "failed to resolve config from PingoraConfig")
		// Update Gateway status to reflect config error and requeue for retry
		if statusErr := r.setConfigErrorStatus(ctx, &gateway, err); statusErr != nil {
			logger.Error(statusErr, "failed to update gateway status")
		}

		return ctrl.Result{RequeueAfter: configErrorRequeueDelay}, nil
	}

	if !gateway.DeletionTimestamp.IsZero() {
		// Gateway is being deleted - nothing special to clean up for Pingora
		return ctrl.Result{}, nil
	}

	if err := r.updateStatus(ctx, &gateway, resolvedConfig); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to update gateway status")
	}

	return ctrl.Result{}, nil
}

//nolint:funlen // status update logic with retry
func (r *PingoraGatewayReconciler) updateStatus(
	ctx context.Context,
	gateway *gatewayv1.Gateway,
	cfg *config.ResolvedPingoraConfig,
) error {
	gatewayKey := types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get fresh copy of the gateway to avoid conflict errors
		var freshGateway gatewayv1.Gateway
		if err := r.Get(ctx, gatewayKey, &freshGateway); err != nil {
			return errors.Wrap(err, "failed to get fresh gateway")
		}

		now := metav1.Now()

		attachedRoutes := r.countAttachedRoutes(ctx, &freshGateway)

		// Set Pingora proxy address as the gateway address
		freshGateway.Status.Addresses = []gatewayv1.GatewayStatusAddress{
			{
				Type:  ptr(gatewayv1.HostnameAddressType),
				Value: cfg.Address,
			},
		}

		freshGateway.Status.Conditions = []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: freshGateway.Generation,
				LastTransitionTime: now,
				Reason:             string(gatewayv1.GatewayReasonAccepted),
				Message:            "Gateway accepted by Pingora controller",
			},
			{
				Type:               string(gatewayv1.GatewayConditionProgrammed),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: freshGateway.Generation,
				LastTransitionTime: now,
				Reason:             string(gatewayv1.GatewayReasonProgrammed),
				Message:            "Gateway programmed in Pingora proxy",
			},
		}

		listenerStatuses := make([]gatewayv1.ListenerStatus, 0, len(freshGateway.Spec.Listeners))

		for _, listener := range freshGateway.Spec.Listeners {
			listenerStatuses = append(listenerStatuses, gatewayv1.ListenerStatus{
				Name: listener.Name,
				SupportedKinds: []gatewayv1.RouteGroupKind{
					{
						Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
						Kind:  "HTTPRoute",
					},
					{
						Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
						Kind:  "GRPCRoute",
					},
				},
				AttachedRoutes: attachedRoutes[listener.Name],
				Conditions: []metav1.Condition{
					{
						Type:               string(gatewayv1.ListenerConditionAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: freshGateway.Generation,
						LastTransitionTime: now,
						Reason:             string(gatewayv1.ListenerReasonAccepted),
						Message:            "Listener accepted",
					},
					{
						Type:               string(gatewayv1.ListenerConditionProgrammed),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: freshGateway.Generation,
						LastTransitionTime: now,
						Reason:             string(gatewayv1.ListenerReasonProgrammed),
						Message:            "Listener programmed",
					},
					{
						Type:               string(gatewayv1.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: freshGateway.Generation,
						LastTransitionTime: now,
						Reason:             string(gatewayv1.ListenerReasonResolvedRefs),
						Message:            "References resolved",
					},
				},
			})
		}

		freshGateway.Status.Listeners = listenerStatuses

		if err := r.Status().Update(ctx, &freshGateway); err != nil {
			return errors.Wrap(err, "failed to update gateway status")
		}

		return nil
	})

	return errors.Wrap(err, "failed to update gateway status after retries")
}

func (r *PingoraGatewayReconciler) setConfigErrorStatus(
	ctx context.Context,
	gateway *gatewayv1.Gateway,
	configErr error,
) error {
	gatewayKey := types.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get fresh copy of the gateway to avoid conflict errors
		var freshGateway gatewayv1.Gateway
		if err := r.Get(ctx, gatewayKey, &freshGateway); err != nil {
			return errors.Wrap(err, "failed to get fresh gateway")
		}

		now := metav1.Now()

		freshGateway.Status.Conditions = []metav1.Condition{
			{
				Type:               string(gatewayv1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: freshGateway.Generation,
				LastTransitionTime: now,
				Reason:             "InvalidParameters",
				Message:            "Failed to resolve PingoraConfig: " + configErr.Error(),
			},
		}

		if err := r.Status().Update(ctx, &freshGateway); err != nil {
			return errors.Wrap(err, "failed to update gateway status")
		}

		return nil
	})

	return errors.Wrap(err, "failed to update gateway status after retries")
}

//nolint:gocognit,gocyclo,cyclop,dupl,funlen // complexity due to counting two route types
func (r *PingoraGatewayReconciler) countAttachedRoutes(
	ctx context.Context,
	gateway *gatewayv1.Gateway,
) map[gatewayv1.SectionName]int32 {
	logger := logging.FromContext(ctx)
	result := make(map[gatewayv1.SectionName]int32)

	for _, listener := range gateway.Spec.Listeners {
		result[listener.Name] = 0
	}

	validator := routebinding.NewValidator(r.Client)

	// Count HTTPRoutes with binding validation
	var httpRouteList gatewayv1.HTTPRouteList

	err := r.List(ctx, &httpRouteList)
	if err != nil {
		logger.Error("failed to list HTTPRoutes for attached routes count", "error", err)
	} else {
		for i := range httpRouteList.Items {
			route := &httpRouteList.Items[i]

			for _, ref := range route.Spec.ParentRefs {
				if !r.refMatchesGateway(ref, gateway, route.Namespace) {
					continue
				}

				routeInfo := &routebinding.RouteInfo{
					Name:        route.Name,
					Namespace:   route.Namespace,
					Hostnames:   route.Spec.Hostnames,
					Kind:        routebinding.KindHTTPRoute,
					SectionName: ref.SectionName,
				}

				bindingResult, bindErr := validator.ValidateBinding(ctx, gateway, routeInfo)
				if bindErr != nil || !bindingResult.Accepted {
					continue
				}

				// Count this route for each matched listener
				for _, listenerName := range bindingResult.MatchedListeners {
					result[listenerName]++
				}
			}
		}
	}

	// Count GRPCRoutes with binding validation
	var grpcRouteList gatewayv1.GRPCRouteList

	err = r.List(ctx, &grpcRouteList)
	if err != nil {
		logger.Error("failed to list GRPCRoutes for attached routes count", "error", err)
	} else {
		for i := range grpcRouteList.Items {
			route := &grpcRouteList.Items[i]

			for _, ref := range route.Spec.ParentRefs {
				if !r.refMatchesGateway(ref, gateway, route.Namespace) {
					continue
				}

				routeInfo := &routebinding.RouteInfo{
					Name:        route.Name,
					Namespace:   route.Namespace,
					Hostnames:   route.Spec.Hostnames,
					Kind:        routebinding.KindGRPCRoute,
					SectionName: ref.SectionName,
				}

				bindingResult, bindErr := validator.ValidateBinding(ctx, gateway, routeInfo)
				if bindErr != nil || !bindingResult.Accepted {
					continue
				}

				// Count this route for each matched listener
				for _, listenerName := range bindingResult.MatchedListeners {
					result[listenerName]++
				}
			}
		}
	}

	return result
}

func (r *PingoraGatewayReconciler) refMatchesGateway(
	ref gatewayv1.ParentReference,
	gateway *gatewayv1.Gateway,
	routeNamespace string,
) bool {
	if string(ref.Name) != gateway.Name {
		return false
	}

	refNamespace := routeNamespace
	if ref.Namespace != nil {
		refNamespace = string(*ref.Namespace)
	}

	return refNamespace == gateway.Namespace
}

// SetupWithManager sets up the controller with the Manager.
func (r *PingoraGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapper := &PingoraConfigMapper{
		Client:           r.Client,
		GatewayClassName: r.GatewayClassName,
		ConfigResolver:   r.ConfigResolver,
	}

	//nolint:wrapcheck // controller-runtime builder pattern
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.Gateway{}).
		// Watch GatewayClass for parametersRef changes
		Watches(
			&gatewayv1.GatewayClass{},
			handler.EnqueueRequestsFromMapFunc(r.gatewayClassToGateways),
		).
		// Watch PingoraConfig for config changes
		Watches(
			&v1alpha1.PingoraConfig{},
			handler.EnqueueRequestsFromMapFunc(mapper.MapConfigToRequests(r.getAllGatewaysForClass)),
		).
		Complete(r)
}

// gatewayClassToGateways maps GatewayClass events to Gateway reconcile requests.
func (r *PingoraGatewayReconciler) gatewayClassToGateways(
	ctx context.Context,
	obj client.Object,
) []reconcile.Request {
	gatewayClass, ok := obj.(*gatewayv1.GatewayClass)
	if !ok {
		return nil
	}

	if gatewayClass.Name != r.GatewayClassName {
		return nil
	}

	return r.getAllGatewaysForClass(ctx)
}

func (r *PingoraGatewayReconciler) getAllGatewaysForClass(ctx context.Context) []reconcile.Request {
	var gatewayList gatewayv1.GatewayList

	err := r.List(ctx, &gatewayList)
	if err != nil {
		return nil
	}

	var requests []reconcile.Request

	for i := range gatewayList.Items {
		gw := &gatewayList.Items[i]
		if string(gw.Spec.GatewayClassName) == r.GatewayClassName {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      gw.Name,
					Namespace: gw.Namespace,
				},
			})
		}
	}

	return requests
}

func ptr[T any](v T) *T {
	return &v
}
