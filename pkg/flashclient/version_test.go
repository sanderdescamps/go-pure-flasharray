package flashclient_test

import (
	"slices"
	"sort"
	"testing"

	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
	"github.com/stretchr/testify/assert"
)

func TestApiVersion(t *testing.T) {

	t.Run("ApiVersion_compare_lessthan", func(t *testing.T) {
		if flashclient.VersionLessThan("v1.2", "1.3") != true {
			t.Errorf("ApiVersionLessThan failed: expected v1.2 to be smaller than 1.3")
		}
		if flashclient.VersionLessThan("1.2", "1.2.1") != true {
			t.Errorf("ApiVersionLessThan failed: expected 1.2 to be smaller than 1.2.1")
		}
		if flashclient.VersionLessThan("1.2.0", "1.2.1") != true {
			t.Errorf("ApiVersionLessThan failed: expected 1.2.0 to be smaller than 1.2.1")
		}
		if flashclient.VersionLessThan("1.2.0", "1.2.0.0") != true {
			t.Errorf("ApiVersionLessThan failed: expected 1.2.0 to be smaller than 1.2.0.0")
		}
		if flashclient.VersionLessThan("1.2", "1.2") != false {
			t.Errorf("ApiVersionLessThan failed: expected 1.2 not to be smaller than 1.2")
		}
		if flashclient.VersionLessThan("1.2.1", "1.2") != false {
			t.Errorf("ApiVersionLessThan failed: expected 1.2.1 not to be smaller than 1.2")
		}
		if flashclient.VersionLessThan("2.1", "1.2") != false {
			t.Errorf("ApiVersionLessThan failed: expected 2.1 not to be smaller than 1.2")
		}
	})

	t.Run("ApiVersion_compare_equal", func(t *testing.T) {
		if flashclient.VersionEqual("v1.2", "1.3") != false {
			t.Errorf("ApiVersionEqual failed: expected v1.2 not to be equal to 1.3")
		}
		if flashclient.VersionEqual("1.2", "1.2.1") != false {
			t.Errorf("ApiVersionEqual failed: expected 1.2 not to be equal to 1.2.1")
		}
		if flashclient.VersionEqual("1.2.0", "1.2.1") != false {
			t.Errorf("ApiVersionEqual failed: expected 1.2.0 not to be equal to 1.2.1")
		}
		if flashclient.VersionEqual("1.2.0", "1.2.0.0") != false {
			t.Errorf("ApiVersionEqual failed: expected 1.2.0 not to be equal to 1.2.0.0")
		}
		if flashclient.VersionEqual("1.2", "1.2") != true {
			t.Errorf("ApiVersionEqual failed: expected 1.2 to be equal to 1.2")
		}
		if flashclient.VersionEqual("1.2.1", "1.2") != false {
			t.Errorf("ApiVersionEqual failed: expected 1.2.1 not to be equal to 1.2")
		}
		if flashclient.VersionEqual("2.1", "1.2") != false {
			t.Errorf("ApiVersionEqual failed: expected 2.1 not to be equal to 1.2")
		}
	})

	t.Run("ApiVersion_sort", func(t *testing.T) {
		versions := flashclient.ApiVersions{
			"1.06", "1.1", "2.4", "2.1", "1.2.1", "1.2.0", "2.2", "2.3", "1.2", "1.2.0.5", "1.2.0",
		}

		correctOrder := flashclient.ApiVersions{
			"1.1", "1.2", "1.2.0", "1.2.0", "1.2.0.5", "1.2.1", "1.06", "2.1", "2.2", "2.3", "2.4",
		}
		sort.Sort(versions)
		if slices.Compare(versions, correctOrder) != 0 {
			t.Errorf("ApiVersion sort failed: expected %v, got %v", correctOrder, versions)
		}
	})

	t.Run("ApiVersion_invalid_version", func(t *testing.T) {
		assert.Panics(t, func() {
			flashclient.VersionLessThan("1.2", "invalid")
		}, "Expected panic for invalid version string")

		assert.Panics(t, func() {
			flashclient.VersionLessThan("invalid", "1.2")
		}, "Expected panic for invalid version string")

		assert.Panics(t, func() {
			flashclient.VersionEqual("invalid2", "invalid1")
		}, "Expected panic for invalid version string")

		assert.Panics(t, func() {
			flashclient.VersionEqual("invalid1", "invalid2")
		}, "Expected panic for invalid version string")

		assert.Panics(t, func() {
			flashclient.VersionEqual("1.2.invalid", "1.2.invalid")
		}, "Expected panic for invalid version string")
	})

	t.Run("ApiVersion_contains", func(t *testing.T) {
		versions := flashclient.ApiVersions{
			"1.06", "1.1", "2.4", "2.1", "1.2.1", "1.2.0", "2.2", "2.3", "1.2",
		}
		assert.True(t, versions.Contains("1.2.0"), "Expected versions to contain 1.2.0")
		assert.False(t, versions.Contains("3.0"), "Expected versions not to contain 3.0")
	})

	t.Run("ApiVersion_latest", func(t *testing.T) {
		versions := flashclient.ApiVersions{
			"1.06", "1.1", "2.4", "2.1", "1.2.1", "1.2.0", "2.2", "2.3", "1.2",
		}
		assert.Equal(t, "2.4", versions.Latest(), "Expected latest version to be 2.4")

		assert.Equal(t, "", flashclient.ApiVersions{}.Latest(), "Expected latest version to be empty for an empty list")
	})
}
