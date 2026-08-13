package flashclient_test

import (
	"testing"
)

func TestArrays(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("arrays_list", func(t *testing.T) {
		al, err := client.GetArrays()
		if err != nil {
			t.Fatalf("Failed to get arrays: %v", err)
		}
		if len(al) < 1 {
			t.Fatalf("Expected at least 1 array")
		}
	})
}
