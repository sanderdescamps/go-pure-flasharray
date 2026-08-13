package flashclient

import "strings"

func CleanURI(u string) string {
	u = strings.TrimSuffix(u, "/")
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		return "https://" + u
	}
	return u
}

func toPtr[T any](v T) *T {
	return &v
}
