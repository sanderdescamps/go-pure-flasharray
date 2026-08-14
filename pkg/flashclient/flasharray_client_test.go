package flashclient_test

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

const (
	DEFAULT_API_ENDPOINT = "http://localhost:8080"
	DEFAULT_API_TOKEN    = "fake-auth-token"
)

func init() {
	if os.Getenv("PUREFA_TEST_ENDPOINT") == "" {
		os.Setenv("PUREFA_TEST_ENDPOINT", DEFAULT_API_ENDPOINT)
	}
	if os.Getenv("PUREFA_TEST_API_TOKEN") == "" {
		os.Setenv("PUREFA_TEST_API_TOKEN", DEFAULT_API_TOKEN)
	}
}

func readEnvs(t *testing.T) (string, string, flashclient.ClientConfig) {
	endpoint := os.Getenv("PUREFA_TEST_ENDPOINT")
	apiToken := os.Getenv("PUREFA_TEST_API_TOKEN")
	if endpoint == "" || apiToken == "" {
		t.Fatalf("PUREFA_TEST_ENDPOINT or PUREFA_TEST_API_TOKEN not set, skipping test")
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
	} else {
		cfg.Debug = false
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

func setupTestClient(t *testing.T) (*flashclient.FAClient, func()) {
	endpoint, apiToken, cfg := readEnvs(t)

	client, err := flashclient.NewRestClient(endpoint, apiToken, cfg)
	if err != nil {
		t.Fatalf("Failed to create new REST client: %v", err)
	}

	return client, func() {
		err := client.Close()
		if err != nil {
			t.Fatalf("Failed to close client, got %v", err)
		}
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
