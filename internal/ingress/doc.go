// Package ingress provides conversion from Gateway API resources
// to Pingora routing configuration via gRPC.
//
// # Overview
//
// The PingoraBuilder type converts Gateway API HTTPRoute and GRPCRoute
// resources into Pingora protobuf route configurations. It handles:
//
//   - Hostname extraction from route.spec.hostnames
//   - Path matching (Exact, PathPrefix, and Regex types)
//   - Header and query parameter matching
//   - Backend service resolution to cluster-internal addresses
//   - Weight-based load balancing configuration
//
// # Service Resolution
//
// Backend references are resolved to fully-qualified cluster DNS names:
//
//	<service>.<namespace>.svc.<cluster-domain>:<port>
//
// # Route Building
//
// The builder creates protobuf messages that are sent to the Pingora proxy
// via gRPC for dynamic route configuration updates.
package ingress
