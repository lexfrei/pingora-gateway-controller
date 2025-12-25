package controller

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/lexfrei/pingora-gateway-controller/api/v1alpha1"
	"github.com/lexfrei/pingora-gateway-controller/internal/config"
	"github.com/lexfrei/pingora-gateway-controller/internal/logging"
	"github.com/lexfrei/pingora-gateway-controller/internal/routebinding"
)

const (
	// Route status messages for Pingora.
	pingoraRouteAcceptedMessage = "Route accepted and programmed in Pingora proxy"

	// startupPendingRequeueDelay is the delay before retrying when startup sync is pending.
	startupPendingRequeueDelay = 1 * time.Second

	// resolvedRefsMessage is the default message for resolved refs condition.
	resolvedRefsMessage = "References resolved"
)

// PingoraHTTPRouteReconciler reconciles HTTPRoute resources and synchronizes them
// to Pingora proxy via gRPC.
//
// Key behaviors:
//   - Watches all HTTPRoute resources in the cluster
//   - Filters routes by parent Gateway's GatewayClass
//   - Uses shared PingoraRouteSyncer for unified sync with GRPCRoutes
//   - Updates Pingora proxy config via gRPC (hot-reload)
//   - Updates HTTPRoute status with acceptance conditions
//
// On startup, the reconciler performs a full sync to ensure Pingora configuration
// matches the current state of route resources.
type PingoraHTTPRouteReconciler struct {
	client.Client

	// Scheme is the runtime scheme for API type registration.
	Scheme *runtime.Scheme

	// GatewayClassName filters which routes to process.
	GatewayClassName string

	// ControllerName is reported in HTTPRoute status.
	ControllerName string

	// RouteSyncer provides unified sync for both HTTP and GRPC routes.
	RouteSyncer *PingoraRouteSyncer

	// bindingValidator validates route binding to Gateway listeners.
	bindingValidator *routebinding.Validator

	// startupComplete indicates whether the startup sync has completed.
	// This prevents race conditions between startup sync and reconcile loop.
	startupComplete atomic.Bool
}

func (r *PingoraHTTPRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Wait for startup sync to complete before processing reconcile events
	// to prevent race conditions with gRPC updates
	if !r.startupComplete.Load() {
		return ctrl.Result{RequeueAfter: startupPendingRequeueDelay}, nil
	}

	ctx = logging.WithReconcileID(ctx)
	logger := logging.Component(ctx, "pingora-httproute-reconciler").With("httproute", req.String())
	ctx = logging.WithLogger(ctx, logger)

	var route gatewayv1.HTTPRoute
	if err := r.Get(ctx, req.NamespacedName, &route); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("httproute deleted, triggering full sync")

			return r.syncAndUpdateStatus(ctx)
		}

		return ctrl.Result{}, errors.Wrap(err, "failed to get httproute")
	}

	if !r.isRouteForOurGateway(ctx, &route) {
		return ctrl.Result{}, nil
	}

	logger.Info("reconciling httproute")

	return r.syncAndUpdateStatus(ctx)
}

func (r *PingoraHTTPRouteReconciler) syncAndUpdateStatus(ctx context.Context) (ctrl.Result, error) {
	logger := logging.FromContext(ctx)

	result, syncResult, syncErr := r.RouteSyncer.SyncAllRoutes(ctx)

	// Update status for all HTTP routes with per-parent binding results
	var statusUpdateErr error

	if syncResult != nil {
		for i := range syncResult.HTTPRoutes {
			route := &syncResult.HTTPRoutes[i]
			routeKey := route.Namespace + "/" + route.Name
			bindingInfo := syncResult.HTTPRouteBindings[routeKey]

			if err := r.updateRouteStatus(ctx, route, bindingInfo, syncErr); err != nil {
				logger.Error("failed to update httproute status", "error", err)
				// Keep first error to return for requeue with backoff
				if statusUpdateErr == nil {
					statusUpdateErr = err
				}
			}
		}
	}

	if syncErr != nil && result.RequeueAfter == 0 {
		// Don't propagate error for non-retriable errors
		return result, nil
	}

	// Return error if status updates failed - controller-runtime will requeue with backoff
	if statusUpdateErr != nil {
		return ctrl.Result{}, statusUpdateErr
	}

	return result, nil
}

func (r *PingoraHTTPRouteReconciler) isRouteForOurGateway(ctx context.Context, route *gatewayv1.HTTPRoute) bool {
	return IsRouteAcceptedByGateway(ctx, r.Client, r.bindingValidator, r.GatewayClassName, HTTPRouteWrapper{route})
}

//nolint:funlen,dupl // status update logic; similar structure to GRPCRoute controller is intentional
func (r *PingoraHTTPRouteReconciler) updateRouteStatus(
	ctx context.Context,
	route *gatewayv1.HTTPRoute,
	bindingInfo routeBindingInfo,
	syncErr error,
) error {
	routeKey := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get fresh copy of the route to avoid conflict errors
		var freshRoute gatewayv1.HTTPRoute
		if err := r.Get(ctx, routeKey, &freshRoute); err != nil {
			return errors.Wrap(err, "failed to get fresh httproute")
		}

		now := metav1.Now()
		freshRoute.Status.Parents = nil

		for refIdx, ref := range freshRoute.Spec.ParentRefs {
			if ref.Kind != nil && *ref.Kind != kindGateway {
				continue
			}

			namespace := freshRoute.Namespace
			if ref.Namespace != nil {
				namespace = string(*ref.Namespace)
			}

			var gateway gatewayv1.Gateway
			if err := r.Get(ctx, client.ObjectKey{Name: string(ref.Name), Namespace: namespace}, &gateway); err != nil {
				continue
			}

			if gateway.Spec.GatewayClassName != gatewayv1.ObjectName(r.GatewayClassName) {
				continue
			}

			// Get binding result for this parent ref
			bindingResult, hasBinding := bindingInfo.bindingResults[refIdx]

			status := metav1.ConditionTrue
			reason := string(gatewayv1.RouteReasonAccepted)
			message := pingoraRouteAcceptedMessage

			if syncErr != nil {
				status = metav1.ConditionFalse
				reason = string(gatewayv1.RouteReasonPending)
				message = syncErr.Error()
			} else if hasBinding && !bindingResult.Accepted {
				status = metav1.ConditionFalse
				reason = string(bindingResult.Reason)
				message = bindingResult.Message
			}

			// Create copy to avoid pointer to loop variable
			parentNS := gatewayv1.Namespace(namespace)

			parentStatus := gatewayv1.RouteParentStatus{
				ParentRef: gatewayv1.ParentReference{
					Group:       ref.Group,
					Kind:        ref.Kind,
					Namespace:   &parentNS,
					Name:        ref.Name,
					SectionName: ref.SectionName,
				},
				ControllerName: gatewayv1.GatewayController(r.ControllerName),
				Conditions: []metav1.Condition{
					{
						Type:               string(gatewayv1.RouteConditionAccepted),
						Status:             status,
						ObservedGeneration: freshRoute.Generation,
						LastTransitionTime: now,
						Reason:             reason,
						Message:            message,
					},
					{
						Type:               string(gatewayv1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: freshRoute.Generation,
						LastTransitionTime: now,
						Reason:             string(gatewayv1.RouteReasonResolvedRefs),
						Message:            resolvedRefsMessage,
					},
				},
			}

			freshRoute.Status.Parents = append(freshRoute.Status.Parents, parentStatus)
		}

		if err := r.Status().Update(ctx, &freshRoute); err != nil {
			return errors.Wrap(err, "failed to update httproute status")
		}

		return nil
	})

	return errors.Wrap(err, "failed to update httproute status after retries")
}

func (r *PingoraHTTPRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.bindingValidator = routebinding.NewValidator(r.Client)

	mapper := &PingoraConfigMapper{
		Client:           r.Client,
		GatewayClassName: r.GatewayClassName,
		ConfigResolver:   r.RouteSyncer.ConfigResolver,
	}

	err := ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1.HTTPRoute{}).
		// Filter out status-only updates to prevent infinite reconciliation loops.
		// We only care about spec changes (generation changes) or deletions.
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Watches(
			&gatewayv1.Gateway{},
			handler.EnqueueRequestsFromMapFunc(r.findRoutesForGateway),
		).
		// Watch PingoraConfig for config changes
		Watches(
			&v1alpha1.PingoraConfig{},
			handler.EnqueueRequestsFromMapFunc(mapper.MapConfigToRequests(r.getAllRelevantRoutes)),
		).
		// Watch Secrets for TLS credential changes
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(mapper.MapSecretToRequests(r.getAllRelevantRoutes)),
		).
		// Watch ReferenceGrant for cross-namespace permission changes
		Watches(
			&gatewayv1beta1.ReferenceGrant{},
			handler.EnqueueRequestsFromMapFunc(r.findRoutesForReferenceGrant),
		).
		Complete(r)
	if err != nil {
		return errors.Wrap(err, "failed to setup pingora httproute controller")
	}

	// Add startup runnable for initial sync
	addErr := mgr.Add(r)
	if addErr != nil {
		return errors.Wrap(addErr, "failed to add startup sync runnable")
	}

	return nil
}

// Start implements manager.Runnable for startup sync.
func (r *PingoraHTTPRouteReconciler) Start(ctx context.Context) error {
	// Mark startup as complete when this function returns,
	// regardless of success or failure
	defer r.startupComplete.Store(true)

	logger := logging.Component(ctx, "pingora-httproute-startup-sync")
	logger.Info("performing startup sync of Pingora configuration")

	ctx = logging.WithLogger(ctx, logger)

	_, err := r.syncAndUpdateStatus(ctx)
	if err != nil {
		logger.Error("startup sync failed", "error", err)
		// Don't return error - allow controller to start even if initial sync fails
	} else {
		logger.Info("startup sync completed successfully")
	}

	return nil
}

func (r *PingoraHTTPRouteReconciler) findRoutesForGateway(
	ctx context.Context,
	obj client.Object,
) []reconcile.Request {
	var routeList gatewayv1.HTTPRouteList
	if err := r.List(ctx, &routeList); err != nil {
		return nil
	}

	routes := make([]Route, len(routeList.Items))
	for i := range routeList.Items {
		routes[i] = HTTPRouteWrapper{&routeList.Items[i]}
	}

	return FindRoutesForGateway(obj, r.GatewayClassName, routes)
}

func (r *PingoraHTTPRouteReconciler) findRoutesForReferenceGrant(
	ctx context.Context,
	obj client.Object,
) []reconcile.Request {
	var routeList gatewayv1.HTTPRouteList

	err := r.List(ctx, &routeList)
	if err != nil {
		return nil
	}

	// Collect routes managed by our Gateway as Route
	routes := make([]Route, 0, len(routeList.Items))

	for i := range routeList.Items {
		route := &routeList.Items[i]
		if r.isRouteForOurGateway(ctx, route) {
			routes = append(routes, HTTPRouteWrapper{route})
		}
	}

	return FindRoutesForReferenceGrant(obj, routes)
}

func (r *PingoraHTTPRouteReconciler) getAllRelevantRoutes(ctx context.Context) []reconcile.Request {
	var routeList gatewayv1.HTTPRouteList

	err := r.List(ctx, &routeList)
	if err != nil {
		return nil
	}

	routes := make([]Route, len(routeList.Items))
	for i := range routeList.Items {
		routes[i] = HTTPRouteWrapper{&routeList.Items[i]}
	}

	return FilterAcceptedRoutes(ctx, r.Client, r.bindingValidator, r.GatewayClassName, routes)
}

// PingoraConfigMapper maps PingoraConfig and Secret changes to route reconcile requests.
type PingoraConfigMapper struct {
	Client           client.Client
	GatewayClassName string
	ConfigResolver   *config.PingoraResolver
}

// MapConfigToRequests returns a function that maps PingoraConfig changes to route requests.
func (m *PingoraConfigMapper) MapConfigToRequests(
	getRoutes func(ctx context.Context) []reconcile.Request,
) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		pingoraConfig, ok := obj.(*v1alpha1.PingoraConfig)
		if !ok {
			return nil
		}

		// Check if this config is referenced by our GatewayClass
		var gatewayClass gatewayv1.GatewayClass
		if err := m.Client.Get(ctx, client.ObjectKey{Name: m.GatewayClassName}, &gatewayClass); err != nil {
			return nil
		}

		if gatewayClass.Spec.ParametersRef == nil {
			return nil
		}

		ref := gatewayClass.Spec.ParametersRef
		if string(ref.Group) != config.PingoraParametersRefGroup ||
			string(ref.Kind) != config.PingoraParametersRefKind ||
			ref.Name != pingoraConfig.Name {
			return nil
		}

		// Config matches, return all relevant routes
		return getRoutes(ctx)
	}
}

// MapSecretToRequests returns a function that maps Secret changes to route requests.
func (m *PingoraConfigMapper) MapSecretToRequests(
	getRoutes func(ctx context.Context) []reconcile.Request,
) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			return nil
		}

		// Get the PingoraConfig for our GatewayClass
		var gatewayClass gatewayv1.GatewayClass
		if err := m.Client.Get(ctx, client.ObjectKey{Name: m.GatewayClassName}, &gatewayClass); err != nil {
			return nil
		}

		pingoraConfig, err := m.ConfigResolver.GetConfigForGatewayClass(ctx, &gatewayClass)
		if err != nil {
			return nil
		}

		// Check if this secret is referenced by the config
		if pingoraConfig.Spec.TLS != nil && pingoraConfig.Spec.TLS.SecretRef != nil {
			secretRef := pingoraConfig.Spec.TLS.SecretRef

			secretNS := secretRef.Namespace
			if secretNS == "" {
				secretNS = "default"
			}

			if secret.Name == secretRef.Name && secret.Namespace == secretNS {
				return getRoutes(ctx)
			}
		}

		return nil
	}
}
