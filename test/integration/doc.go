//go:build integration

// Package integration provides integration and E2E tests for the
// pingora-gateway-controller. These tests require a running container
// runtime (Docker/Podman) and verify communication between Go controller
// and Rust Pingora proxy via gRPC.
//
// Run with: go test -v -tags=integration -timeout=10m ./test/integration/...
//
// Environment variables:
//   - PINGORA_PROXY_IMAGE: Pre-built image name (default: builds from Containerfile)
package integration
