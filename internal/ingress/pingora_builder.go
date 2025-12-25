package ingress

import (
	"fmt"
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

// parseGatewayDuration parses a Gateway API duration string (e.g., "10s", "1m").
//
//nolint:wrapcheck // standard library errors are descriptive
func parseGatewayDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// PingoraBuilder builds Pingora route configurations from Gateway API resources.
type PingoraBuilder struct {
	clusterDomain string
}

// NewPingoraBuilder creates a new PingoraBuilder.
func NewPingoraBuilder(clusterDomain string) *PingoraBuilder {
	return &PingoraBuilder{
		clusterDomain: clusterDomain,
	}
}

// BuildHTTPRoute converts a Gateway API HTTPRoute to a Pingora HTTPRoute.
//
//nolint:dupl // HTTPRoute and GRPCRoute have similar structure but different types
func (b *PingoraBuilder) BuildHTTPRoute(route *gatewayv1.HTTPRoute) *routingv1.HTTPRoute {
	result := &routingv1.HTTPRoute{
		Id:        fmt.Sprintf("%s/%s", route.Namespace, route.Name),
		Hostnames: make([]string, 0, len(route.Spec.Hostnames)),
		Rules:     make([]*routingv1.HTTPRouteRule, 0, len(route.Spec.Rules)),
	}

	// Convert hostnames
	for _, hostname := range route.Spec.Hostnames {
		result.Hostnames = append(result.Hostnames, string(hostname))
	}

	// Convert rules
	for _, rule := range route.Spec.Rules {
		result.Rules = append(result.Rules, b.buildHTTPRouteRule(route.Namespace, &rule))
	}

	return result
}

// BuildGRPCRoute converts a Gateway API GRPCRoute to a Pingora GRPCRoute.
//
//nolint:dupl // GRPCRoute and HTTPRoute have similar structure but different types
func (b *PingoraBuilder) BuildGRPCRoute(route *gatewayv1.GRPCRoute) *routingv1.GRPCRoute {
	result := &routingv1.GRPCRoute{
		Id:        fmt.Sprintf("%s/%s", route.Namespace, route.Name),
		Hostnames: make([]string, 0, len(route.Spec.Hostnames)),
		Rules:     make([]*routingv1.GRPCRouteRule, 0, len(route.Spec.Rules)),
	}

	// Convert hostnames
	for _, hostname := range route.Spec.Hostnames {
		result.Hostnames = append(result.Hostnames, string(hostname))
	}

	// Convert rules
	for _, rule := range route.Spec.Rules {
		result.Rules = append(result.Rules, b.buildGRPCRouteRule(route.Namespace, &rule))
	}

	return result
}

func (b *PingoraBuilder) buildHTTPRouteRule(namespace string, rule *gatewayv1.HTTPRouteRule) *routingv1.HTTPRouteRule {
	result := &routingv1.HTTPRouteRule{
		Matches:  make([]*routingv1.HTTPRouteMatch, 0),
		Backends: make([]*routingv1.Backend, 0),
	}

	// Convert matches
	if len(rule.Matches) == 0 {
		// Default match: all paths
		result.Matches = append(result.Matches, &routingv1.HTTPRouteMatch{
			Path: &routingv1.PathMatch{
				Type:  routingv1.PathMatchType_PATH_MATCH_TYPE_PREFIX,
				Value: "/",
			},
		})
	} else {
		for _, match := range rule.Matches {
			result.Matches = append(result.Matches, b.buildHTTPRouteMatch(&match))
		}
	}

	// Convert backend references
	for _, backendRef := range rule.BackendRefs {
		backend := b.buildBackend(namespace, &backendRef.BackendRef)
		if backend != nil {
			result.Backends = append(result.Backends, backend)
		}
	}

	// Convert timeouts
	if rule.Timeouts != nil && rule.Timeouts.Request != nil {
		timeout, err := parseGatewayDuration(string(*rule.Timeouts.Request))
		if err == nil {
			ms := timeout.Milliseconds()
			if ms > 0 {
				result.TimeoutMs = uint64(ms)
			}
		}
	}

	return result
}

func (b *PingoraBuilder) buildHTTPRouteMatch(match *gatewayv1.HTTPRouteMatch) *routingv1.HTTPRouteMatch {
	result := &routingv1.HTTPRouteMatch{
		Headers:     make([]*routingv1.HeaderMatch, 0),
		QueryParams: make([]*routingv1.QueryParamMatch, 0),
	}

	// Convert path match
	if match.Path != nil {
		result.Path = &routingv1.PathMatch{
			Value: *match.Path.Value,
		}
		switch *match.Path.Type {
		case gatewayv1.PathMatchExact:
			result.Path.Type = routingv1.PathMatchType_PATH_MATCH_TYPE_EXACT
		case gatewayv1.PathMatchPathPrefix:
			result.Path.Type = routingv1.PathMatchType_PATH_MATCH_TYPE_PREFIX
		case gatewayv1.PathMatchRegularExpression:
			result.Path.Type = routingv1.PathMatchType_PATH_MATCH_TYPE_REGEX
		}
	}

	// Convert method
	if match.Method != nil {
		result.Method = string(*match.Method)
	}

	// Convert headers
	for _, header := range match.Headers {
		result.Headers = append(result.Headers, b.buildHeaderMatch(&header))
	}

	// Convert query params
	for _, qp := range match.QueryParams {
		result.QueryParams = append(result.QueryParams, b.buildQueryParamMatch(&qp))
	}

	return result
}

func (b *PingoraBuilder) buildHeaderMatch(match *gatewayv1.HTTPHeaderMatch) *routingv1.HeaderMatch {
	result := &routingv1.HeaderMatch{
		Name:  string(match.Name),
		Value: match.Value,
	}

	if match.Type != nil {
		switch *match.Type {
		case gatewayv1.HeaderMatchExact:
			result.Type = routingv1.HeaderMatchType_HEADER_MATCH_TYPE_EXACT
		case gatewayv1.HeaderMatchRegularExpression:
			result.Type = routingv1.HeaderMatchType_HEADER_MATCH_TYPE_REGEX
		}
	} else {
		result.Type = routingv1.HeaderMatchType_HEADER_MATCH_TYPE_EXACT
	}

	return result
}

func (b *PingoraBuilder) buildQueryParamMatch(match *gatewayv1.HTTPQueryParamMatch) *routingv1.QueryParamMatch {
	result := &routingv1.QueryParamMatch{
		Name:  string(match.Name),
		Value: match.Value,
	}

	if match.Type != nil {
		switch *match.Type {
		case gatewayv1.QueryParamMatchExact:
			result.Type = routingv1.QueryParamMatchType_QUERY_PARAM_MATCH_TYPE_EXACT
		case gatewayv1.QueryParamMatchRegularExpression:
			result.Type = routingv1.QueryParamMatchType_QUERY_PARAM_MATCH_TYPE_REGEX
		}
	} else {
		result.Type = routingv1.QueryParamMatchType_QUERY_PARAM_MATCH_TYPE_EXACT
	}

	return result
}

func (b *PingoraBuilder) buildGRPCRouteRule(namespace string, rule *gatewayv1.GRPCRouteRule) *routingv1.GRPCRouteRule {
	result := &routingv1.GRPCRouteRule{
		Matches:  make([]*routingv1.GRPCRouteMatch, 0),
		Backends: make([]*routingv1.Backend, 0),
	}

	// Convert matches
	for _, match := range rule.Matches {
		result.Matches = append(result.Matches, b.buildGRPCRouteMatch(&match))
	}

	// Convert backend references
	for _, backendRef := range rule.BackendRefs {
		backend := b.buildBackend(namespace, &backendRef.BackendRef)
		if backend != nil {
			result.Backends = append(result.Backends, backend)
		}
	}

	return result
}

func (b *PingoraBuilder) buildGRPCRouteMatch(match *gatewayv1.GRPCRouteMatch) *routingv1.GRPCRouteMatch {
	result := &routingv1.GRPCRouteMatch{
		Headers: make([]*routingv1.HeaderMatch, 0),
	}

	// Convert method match
	if match.Method != nil {
		result.Method = &routingv1.GRPCMethodMatch{}

		if match.Method.Service != nil {
			result.Method.Service = *match.Method.Service
		}

		if match.Method.Method != nil {
			result.Method.Method = *match.Method.Method
		}

		if match.Method.Type != nil {
			switch *match.Method.Type {
			case gatewayv1.GRPCMethodMatchExact:
				result.Method.Type = routingv1.GRPCMethodMatchType_GRPC_METHOD_MATCH_TYPE_EXACT
			case gatewayv1.GRPCMethodMatchRegularExpression:
				result.Method.Type = routingv1.GRPCMethodMatchType_GRPC_METHOD_MATCH_TYPE_REGEX
			}
		} else {
			result.Method.Type = routingv1.GRPCMethodMatchType_GRPC_METHOD_MATCH_TYPE_EXACT
		}
	}

	// Convert headers
	for _, header := range match.Headers {
		result.Headers = append(result.Headers, b.buildGRPCHeaderMatch(&header))
	}

	return result
}

func (b *PingoraBuilder) buildGRPCHeaderMatch(match *gatewayv1.GRPCHeaderMatch) *routingv1.HeaderMatch {
	result := &routingv1.HeaderMatch{
		Name:  string(match.Name),
		Value: match.Value,
	}

	if match.Type != nil {
		switch *match.Type {
		case gatewayv1.GRPCHeaderMatchExact:
			result.Type = routingv1.HeaderMatchType_HEADER_MATCH_TYPE_EXACT
		case gatewayv1.GRPCHeaderMatchRegularExpression:
			result.Type = routingv1.HeaderMatchType_HEADER_MATCH_TYPE_REGEX
		}
	} else {
		result.Type = routingv1.HeaderMatchType_HEADER_MATCH_TYPE_EXACT
	}

	return result
}

func (b *PingoraBuilder) buildBackend(namespace string, ref *gatewayv1.BackendRef) *routingv1.Backend {
	// Only support Service backends
	if ref.Kind != nil && *ref.Kind != "Service" {
		return nil
	}

	// Determine namespace
	backendNamespace := namespace
	if ref.Namespace != nil {
		backendNamespace = string(*ref.Namespace)
	}

	// Build service address
	address := fmt.Sprintf("%s.%s.svc.%s:%d",
		string(ref.Name),
		backendNamespace,
		b.clusterDomain,
		*ref.Port,
	)

	result := &routingv1.Backend{
		Address:  address,
		Weight:   1,
		Protocol: routingv1.BackendProtocol_BACKEND_PROTOCOL_HTTP,
	}

	// Set weight if specified
	if ref.Weight != nil && *ref.Weight > 0 {
		result.Weight = uint32(*ref.Weight)
	}

	return result
}
