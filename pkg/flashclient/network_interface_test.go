package flashclient_test

import (
	"testing"
)

func TestNetworkInterfaces(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("network_interfaces_1", func(t *testing.T) {
		interfaces, err := client.GetNetworkInterfaces()
		if err != nil {
			t.Fatalf("Failed to get network interfaces: %v", err)
		}
		if lenInterfaces := len(interfaces); lenInterfaces > 0 {
			t.Logf("%d Network interfaces found", lenInterfaces)
		} else {
			t.Errorf("Expected at least 1 network interface")
		}
	})
}
