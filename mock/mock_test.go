package mock_test

import (
	"strings"
	"testing"

	"github.com/sanderdescamps/go-purefa-mock/mock"
)

func TestMock(t *testing.T) {
	array, err := mock.NewArrayFromDir("./../test-data")
	if err != nil {
		t.Fatalf("Failed to load mock array from directory: %v", err)
	}

	t.Run("list-versions", func(t *testing.T) {
		t.Logf("array versions: %s", strings.Join(array.Versions, ", "))
	})

	t.Run("start-mock", func(t *testing.T) {
		server := mock.NewMock(array)
		server.Start("localhost", 8080)
		server.Stop()
	})

}
