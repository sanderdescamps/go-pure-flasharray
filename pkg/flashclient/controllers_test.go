package flashclient_test

import (
	"testing"
)

func TestControllers(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("controller_list", func(t *testing.T) {
		controllers, err := client.GetControllers()
		if err != nil {
			t.Fatalf("Failed to get controllers: %v", err)
		}
		if len(controllers) < 2 {
			t.Fatalf("Expected at least 2 controllers")
		}
	})
}
