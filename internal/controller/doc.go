// Package controller implements Kubernetes controllers for Gateway API resources.
//
// The package provides controllers for the Pingora Gateway Controller:
//
//   - PingoraGatewayReconciler: Watches Gateway resources and updates status
//     based on Pingora proxy connectivity.
//
//   - PingoraHTTPRouteReconciler: Watches HTTPRoute resources and synchronizes
//     them to Pingora proxy via gRPC.
//
//   - PingoraGRPCRouteReconciler: Watches GRPCRoute resources and synchronizes
//     them to Pingora proxy via gRPC.
//
// # Architecture
//
// The controllers follow the standard controller-runtime reconciliation pattern:
//
//	┌─────────────┐    watch     ┌─────────────────────────┐
//	│ HTTPRoute   │─────────────>│ PingoraHTTPRouteReconciler│
//	│ resources   │              │                         │
//	└─────────────┘              └───────────┬─────────────┘
//	                                         │
//	┌─────────────┐    watch                 │ gRPC
//	│ GRPCRoute   │─────────────>│           │
//	│ resources   │              │           ▼
//	└─────────────┘              │  ┌─────────────────┐
//	       │                     │  │ Pingora Proxy   │
//	       │                     │  │ (route config)  │
//	       ▼                     │  └─────────────────┘
//	┌─────────────────────────┐  │
//	│ PingoraGRPCRouteReconciler│ │
//	└─────────────────────────┘  │
//
// # Configuration
//
// Controllers are configured via PingoraConfig CRD which specifies the gRPC
// endpoint address and connection parameters for the Pingora proxy.
//
// # Leader Election
//
// When running multiple replicas for high availability, enable leader election
// via --leader-elect flag to ensure only one controller actively reconciles
// resources at a time.
package controller
