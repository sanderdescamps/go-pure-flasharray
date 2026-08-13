package flashclient_test

import (
	"testing"
)

func TestDrives(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("drives_list", func(t *testing.T) {
		drives, err := client.GetDrives()
		if err != nil {
			t.Fatalf("Failed to get drives: %v", err)
		}
		if len(drives) < 1 {
			t.Fatalf("Expected at least 1 drive")
		}
	})
}
