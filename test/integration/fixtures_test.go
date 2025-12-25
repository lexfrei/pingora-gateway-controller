//go:build integration

package integration

import (
	routingv1 "github.com/lexfrei/pingora-gateway-controller/pkg/api/routing/v1"
)

// NewHTTPRoute creates a test HTTPRoute with a simple path prefix match.
func NewHTTPRoute(id string, hostnames []string, pathPrefix, backendAddr string) *routingv1.HTTPRoute {
	return &routingv1.HTTPRoute{
		Id:        id,
		Hostnames: hostnames,
		Rules: []*routingv1.HTTPRouteRule{
			{
				Matches: []*routingv1.HTTPRouteMatch{
					{
						Path: &routingv1.PathMatch{
							Type:  routingv1.PathMatchType_PATH_MATCH_TYPE_PREFIX,
							Value: pathPrefix,
						},
					},
				},
				Backends: []*routingv1.Backend{
					{
						Address:  backendAddr,
						Weight:   1,
						Protocol: routingv1.BackendProtocol_BACKEND_PROTOCOL_HTTP,
					},
				},
			},
		},
	}
}

// NewHTTPRouteExact creates a test HTTPRoute with exact path match.
func NewHTTPRouteExact(id string, hostnames []string, exactPath, backendAddr string) *routingv1.HTTPRoute {
	return &routingv1.HTTPRoute{
		Id:        id,
		Hostnames: hostnames,
		Rules: []*routingv1.HTTPRouteRule{
			{
				Matches: []*routingv1.HTTPRouteMatch{
					{
						Path: &routingv1.PathMatch{
							Type:  routingv1.PathMatchType_PATH_MATCH_TYPE_EXACT,
							Value: exactPath,
						},
					},
				},
				Backends: []*routingv1.Backend{
					{
						Address:  backendAddr,
						Weight:   1,
						Protocol: routingv1.BackendProtocol_BACKEND_PROTOCOL_HTTP,
					},
				},
			},
		},
	}
}

// NewGRPCRoute creates a test GRPCRoute.
func NewGRPCRoute(id string, hostnames []string, service, method, backendAddr string) *routingv1.GRPCRoute {
	return &routingv1.GRPCRoute{
		Id:        id,
		Hostnames: hostnames,
		Rules: []*routingv1.GRPCRouteRule{
			{
				Matches: []*routingv1.GRPCRouteMatch{
					{
						Method: &routingv1.GRPCMethodMatch{
							Type:    routingv1.GRPCMethodMatchType_GRPC_METHOD_MATCH_TYPE_EXACT,
							Service: service,
							Method:  method,
						},
					},
				},
				Backends: []*routingv1.Backend{
					{
						Address:  backendAddr,
						Weight:   1,
						Protocol: routingv1.BackendProtocol_BACKEND_PROTOCOL_H2C,
					},
				},
			},
		},
	}
}

// NewBackend creates a test Backend.
func NewBackend(address string, weight uint32) *routingv1.Backend {
	return &routingv1.Backend{
		Address:  address,
		Weight:   weight,
		Protocol: routingv1.BackendProtocol_BACKEND_PROTOCOL_HTTP,
	}
}
