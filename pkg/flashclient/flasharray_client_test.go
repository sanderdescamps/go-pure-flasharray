package flashclient_test

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/mock"
)

const (
	MOCK_API_ENDPOINT = "http://localhost:8080"
	MOCK_API_TOKEN    = "fake-auth-token"
)

var (
	mockServer    *mock.Mock
	clientCounter int32
	apiClient     *flashclient.FAClient
	setupOnce     sync.Once
)

func init() {
	os.Setenv("PUREFA_TEST_MOCK", "false")
	os.Setenv("PUREFA_TEST_ENDPOINT", MOCK_API_ENDPOINT)
	os.Setenv("PUREFA_TEST_API_TOKEN", MOCK_API_TOKEN)
	os.Setenv("PUREFA_TEST_INSECURE", "true")
}

func readEnvs(t *testing.T) (string, string, flashclient.ClientConfig) {
	endpoint := os.Getenv("PUREFA_TEST_ENDPOINT")
	apiToken := os.Getenv("PUREFA_TEST_API_TOKEN")
	if endpoint == "" || apiToken == "" {
		endpoint = MOCK_API_ENDPOINT
		apiToken = MOCK_API_TOKEN
		t.Logf("PUREFA_TEST_ENDPOINT or PUREFA_TEST_API_TOKEN not set, using mock server with endpoint=%s and api-token=%s", endpoint, apiToken)
		os.Setenv("PUREFA_TEST_MOCK", "true")
	}

	cfg := flashclient.DefaultClientConfig()

	if insecure, err := strconv.ParseBool(os.Getenv("PUREFA_TEST_INSECURE")); err == nil && insecure {
		cfg.Insecure = true
	}

	if apiVersion := os.Getenv("PUREFA_TEST_API_VERSION"); apiVersion != "" {
		cfg.ApiVersion = apiVersion
	}

	if debug, err := strconv.ParseBool(os.Getenv("PUREFA_TEST_DEBUG")); err == nil && debug {
		cfg.Debug = true
	}

	if userAgent := os.Getenv("PUREFA_TEST_USER_AGENT"); userAgent != "" {
		cfg.UserAgent = userAgent
	}

	if requestID := os.Getenv("PUREFA_TEST_REQUEST_ID"); requestID != "" {
		cfg.RequestID = requestID
	}

	return endpoint, apiToken, cfg
}

func parseEndpoint(endpoint string) (string, int, error) {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", 0, err
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()

	if host == "" {
		return "", 0, fmt.Errorf("invalid endpoint: %s", endpoint)
	}

	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	return host, portInt, nil
}

func setupTest(t *testing.T) {
	setupOnce.Do(func() {
		endpoint, apiToken, cfg := readEnvs(t)
		if useMock, err := strconv.ParseBool(os.Getenv("PUREFA_TEST_MOCK")); err != nil || useMock {
			t.Log("PUREFA_TEST_MOCK is set to true, setting up mock server...")

			mock, err := mock.NewMockWithTestData(apiToken)
			if err != nil {
				t.Fatalf("Failed to create mock server: %v", err)
			}
			mockServer = mock
			arrayHostname, port, err := parseEndpoint(endpoint)
			if err != nil {
				t.Fatalf("Failed to parse endpoint: %v", err)
			}

			go mockServer.Start(arrayHostname, port)
		}

		client, err := flashclient.NewRestClient(endpoint, apiToken, cfg)
		if err != nil {
			t.Fatalf("Failed to create new REST client: %v", err)
		}
		apiClient = client
	})
	atomic.AddInt32(&clientCounter, 1)
}

func teardownTest() {
	// Decrement the counter when a test completes
	atomic.AddInt32(&clientCounter, -1)

	// Only close the server when all tests have completed
	if mockServer != nil && atomic.LoadInt32(&clientCounter) == 0 {
		mockServer.Stop()
	}
}

func setupTestClient(t *testing.T) (*flashclient.FAClient, func()) {
	setupTest(t)
	return apiClient, func() {
		teardownTest()
	}
}

func TestNewRestClient(t *testing.T) {
	t.Run("login", func(t *testing.T) {
		endpoint, apiToken, cfg := readEnvs(t)
		client, err := flashclient.NewRestClient(endpoint, apiToken, cfg)
		if err != nil {
			t.Fatalf("Failed to create new REST client, got %v", err)
		}

		err = client.RefreshSession()
		if err != nil {
			t.Fatalf("Failed to refresh session, got %v", err)
		}

		versions, err := client.GetVersions()
		if err != nil {
			t.Fatalf("Failed to get client version, got %v", err)
		} else if len(versions) == 0 {
			t.Fatalf("Expected at least one version, got 0")
		}
		t.Logf("Client versions: %s", strings.Join(versions, ", "))

		latest := versions.Latest()
		if latest == "" {
			t.Fatalf("Expected latest version to be non-empty")
		}
		t.Logf("Latest client version: %s", latest)

		err = client.Close()
		if err != nil {
			t.Fatalf("Failed to close client, got %v", err)
		}
	})

	t.Run("debug-client", func(t *testing.T) {
		endpoint, apiToken, cfg := readEnvs(t)
		client, err := flashclient.NewRestClient(endpoint, apiToken, cfg)
		if err != nil {
			t.Fatalf("Failed to create new REST client, got %v", err)
		}

		err = client.Close()
		if err != nil {
			t.Fatalf("Failed to close client, got %v", err)
		}
	})

	t.Run("latest-version", func(t *testing.T) {
		endpoint, _, cfg := readEnvs(t)
		versions, err := flashclient.GetAPISupportedVersions(endpoint, cfg.Insecure)
		if err != nil {
			t.Fatalf("Failed to get API supported versions, got %v", err)
		}
		latest := versions.Latest()
		if latest == "" {
			t.Errorf("Expected a non-empty latest version")
		} else {
			t.Logf("latest api version: %s", latest)
		}

	})
}
