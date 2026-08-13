package fakearray

import "slices"

// WithID returns a filter that matches elements by their FixedReference ID.
// T must embed flashclient.FixedReference (or otherwise expose GetId via a pointer receiver).
func WithIDs[T any, PT interface {
	*T
	GetId() string
}](ids ...string) func(*T) bool {
	return func(c *T) bool {
		return slices.Contains(ids, PT(c).GetId())
	}
}

// WithNames returns a filter that matches elements by their FixedReference name.
// T must embed flashclient.FixedReference or flashclient.NoIdReference (or otherwise expose GetName via a pointer receiver).
func WithNames[T any, PT interface {
	*T
	GetName() string
}](names ...string) func(*T) bool {
	return func(c *T) bool {
		return slices.Contains(names, PT(c).GetName())
	}
}
