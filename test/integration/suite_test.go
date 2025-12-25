//go:build integration

package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

const (
	pingoraImageName = "pingora-proxy"
	pingoraImageTag  = "integration-test"
)

//nolint:gochecknoglobals // Required for TestMain setup shared across tests
var (
	pingoraImage string
	buildOnce    sync.Once
	errBuild     error
)

// TestMain sets up the test environment, including building the Pingora image.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build image once before all tests
	buildOnce.Do(func() {
		pingoraImage, errBuild = getPingoraImage(ctx)
	})

	if errBuild != nil {
		log.Fatalf("Failed to get Pingora image: %v", errBuild)
	}

	log.Printf("Using Pingora image: %s", pingoraImage)

	code := m.Run()
	os.Exit(code)
}

// getPingoraImage returns the Pingora proxy image name.
// If PINGORA_PROXY_IMAGE is set, uses that image.
// Otherwise, builds from proxy/Containerfile.
func getPingoraImage(ctx context.Context) (string, error) {
	// Check if pre-built image is specified
	if img := os.Getenv("PINGORA_PROXY_IMAGE"); img != "" {
		log.Printf("Using pre-built image from PINGORA_PROXY_IMAGE: %s", img)
		return img, nil
	}

	// Build from Containerfile
	log.Println("Building Pingora image from proxy/Containerfile...")

	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	proxyDir := filepath.Join(projectRoot, "proxy")
	imageName := pingoraImageName + ":" + pingoraImageTag

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       proxyDir,
				Dockerfile:    "Containerfile",
				PrintBuildLog: true,
				Repo:          pingoraImageName,
				Tag:           pingoraImageTag,
			},
		},
		Started: false,
	}

	// Create container (which triggers the build)
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// We don't need the container, just the built image
	terminateErr := container.Terminate(ctx)
	if terminateErr != nil {
		log.Printf("Warning: failed to terminate build container: %v", terminateErr)
	}

	log.Printf("Successfully built image: %s", imageName)
	return imageName, nil
}

// findProjectRoot walks up from the current directory to find the project root.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		// Check for go.mod as marker for project root
		_, statErr := os.Stat(filepath.Join(dir, "go.mod"))
		if statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}
