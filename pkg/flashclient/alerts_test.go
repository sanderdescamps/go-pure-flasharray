package flashclient_test

import (
	"testing"
)

func TestAlerts(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("alerts_list", func(t *testing.T) {
		alerts, err := client.GetAlerts()
		if err != nil {
			t.Fatalf("Failed to get alerts: %v", err)
		}
		if len(alerts) < 1 {
			t.Fatalf("Expected at least 1 alert")
		}
	})
}
