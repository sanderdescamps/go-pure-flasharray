package flashclient_test

import (
	"testing"
)

func TestPorts(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("ports_list", func(t *testing.T) {
		ports, err := client.GetPorts()
		if err != nil {
			t.Fatalf("Failed to get ports: %v", err)
		}
		if len(ports) < 1 {
			t.Fatalf("Expected at least 1 port")
		}
	})
}
