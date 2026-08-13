package mock_test

import (
	"strings"
	"testing"
	"time"

	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/mock"
)

func TestMock(t *testing.T) {

	array, err := fakearray.NewArrayWithTestData()
	if err != nil {
		t.Fatalf("Failed to load mock array from directory: %v", err)
	}

	t.Run("list-versions", func(t *testing.T) {
		t.Logf("array versions: %s", strings.Join(array.GetVersions(), ", "))
	})

	t.Run("start-mock", func(t *testing.T) {
		server := mock.NewMock(array)
		errCh := make(chan error)
		go func() {
			errCh <- server.Start("localhost", 8080)
		}()
		time.Sleep(5 * time.Second) // Give the server a moment to start
		server.Stop()
		if err := <-errCh; err != nil {
			t.Fatalf("Failed to start mock server: %v", err)
		}
	})

}
