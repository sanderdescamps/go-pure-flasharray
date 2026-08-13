package flashclient_test

import (
	"testing"
)

func TestHardware(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("hardware_list", func(t *testing.T) {
		hardware, err := client.GetHardware()
		if err != nil {
			t.Fatalf("Failed to get hardware: %v", err)
		}
		if len(hardware) < 1 {
			t.Fatalf("Expected at least 1 hardware item")
		}
	})
}
