package flashclient_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
	"github.com/stretchr/testify/assert"
)

func toPtr[T any](v T) *T {
	return &v
}
func NewString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}

func NewRandomName(prefix string, n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return fmt.Sprintf("%s_%s", prefix, string(b))
}

func TestHelper(t *testing.T) {

	t.Run("clean_uri", func(t *testing.T) {
		values := []struct {
			input    string
			expected string
		}{
			{"test.example.com", "https://test.example.com"},
			{"https://test.example.com/", "https://test.example.com"},
			{"http://test.example.com/", "http://test.example.com"},
			{"test.example.com/", "https://test.example.com"},
		}

		for idx, v := range values {
			t.Run(fmt.Sprintf("clean_uri_%d", idx), func(t *testing.T) {
				result := flashclient.CleanURI(v.input)
				assert.Equalf(t, v.expected, result, "clean_uri failed for input: %s, expected: %s, got: %s", v.input, v.expected, result)
			})
		}
	})
}
