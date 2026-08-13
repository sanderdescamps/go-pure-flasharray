package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func TestResults(t *testing.T) {
	t.Run("results_new", func(t *testing.T) {
		hostShorts := []flashclient.HostShort{
			{
				NoIdReference: flashclient.NoIdReference{
					Name: "Host 1",
				},
			}, {
				NoIdReference: flashclient.NoIdReference{
					Name: "Host 2",
				},
			},
		}
		hostResult := flashclient.NewResults(hostShorts)
		if hostResult.TotalItemCount != 2 {
			t.Errorf("Expected 2 hosts, got %d", hostResult.TotalItemCount)
		}
	})
}
