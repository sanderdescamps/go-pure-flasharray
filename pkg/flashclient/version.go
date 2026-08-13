package flashclient

import (
	"sort"
	"strconv"
	"strings"
)

type ApiVersions []string

type VersionsResponse struct {
	Versions ApiVersions `json:"version"`
}

func (v ApiVersions) Latest() string {
	if len(v) == 0 {
		return ""
	}
	sort.Sort(v)
	return v[len(v)-1]
}

// Len returns the number of elements in the ApiVersions slice.
func (v ApiVersions) Len() int {
	return len(v)
}

// Swap exchanges the elements at indices i and j in the ApiVersions slice.
func (v ApiVersions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less compares two version strings and returns true if the version at index i is less than the version at index j.
func (v ApiVersions) Less(i, j int) bool {
	return VersionLessThan(v[i], v[j])
}

// Contains checks if the given version string is present in the ApiVersions slice.
func (v ApiVersions) Contains(s string) bool {
	for _, i := range v {
		if VersionEqual(i, s) {
			return true
		}
	}
	return false
}

func VersionLessThan(v1, v2 string) bool {
	v1 = strings.ToLower(v1)
	v2 = strings.ToLower(v2)

	v1, _ = strings.CutPrefix(v1, "v")
	v2, _ = strings.CutPrefix(v2, "v")

	split1 := strings.Split(v1, ".")
	split2 := strings.Split(v2, ".")

	shortest := len(split1)
	if len(split2) < shortest {
		shortest = len(split2)
	}

	for i := 0; i < shortest; i++ {
		n1, err1 := strconv.Atoi(split1[i])
		n2, err2 := strconv.Atoi(split2[i])
		if err1 != nil || err2 != nil {
			panic("invalid version number")
		} else if n1 != n2 {
			return n1 < n2
		}
	}

	if len(split1) < len(split2) {
		return true
	}

	return false
}

func VersionEqual(v1, v2 string) bool {
	v1 = strings.ToLower(v1)
	v2 = strings.ToLower(v2)

	v1, _ = strings.CutPrefix(v1, "v")
	v2, _ = strings.CutPrefix(v2, "v")

	split1 := strings.Split(v1, ".")
	split2 := strings.Split(v2, ".")

	if len(split1) != len(split2) {
		return false
	}

	for i := 0; i < len(split1); i++ {
		n1, err1 := strconv.Atoi(split1[i])
		n2, err2 := strconv.Atoi(split2[i])
		if err1 != nil || err2 != nil {
			panic("invalid version number")
		} else if n1 != n2 {
			return false
		}
	}

	return true
}
