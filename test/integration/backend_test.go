//go:build integration

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

// RecordedRequest captures details of an HTTP request for verification.
type RecordedRequest struct {
	Method  string
	Path    string
	Host    string
	Headers http.Header
	Body    []byte
}

// MockBackend is a simple HTTP server that records incoming requests.
type MockBackend struct {
	server   *httptest.Server
	requests []RecordedRequest
	mu       sync.Mutex
}

// StartMockBackend creates and starts a new mock backend server.
// The server responds with 200 OK and echoes request details as JSON.
func StartMockBackend() *MockBackend {
	mb := &MockBackend{
		requests: make([]RecordedRequest, 0),
	}

	mb.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		req := RecordedRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Host:    r.Host,
			Headers: r.Header.Clone(),
			Body:    body,
		}

		mb.mu.Lock()
		mb.requests = append(mb.requests, req)
		mb.mu.Unlock()

		// Respond with request details
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Backend-Received", "true")
		w.WriteHeader(http.StatusOK)

		response := map[string]any{
			"path":   r.URL.Path,
			"method": r.Method,
			"host":   r.Host,
		}

		json.NewEncoder(w).Encode(response)
	}))

	return mb
}

// URL returns the backend server URL.
func (m *MockBackend) URL() string {
	return m.server.URL
}

// GetRequests returns all recorded requests.
func (m *MockBackend) GetRequests() []RecordedRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]RecordedRequest, len(m.requests))
	copy(result, m.requests)

	return result
}

// Reset clears all recorded requests.
func (m *MockBackend) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = m.requests[:0]
}

// Close shuts down the backend server.
func (m *MockBackend) Close() {
	m.server.Close()
}

// LastRequest returns the most recent request, or nil if none.
func (m *MockBackend) LastRequest() *RecordedRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.requests) == 0 {
		return nil
	}

	req := m.requests[len(m.requests)-1]

	return &req
}

// RequestCount returns the number of recorded requests.
func (m *MockBackend) RequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.requests)
}
