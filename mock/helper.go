package mock

import (
	"encoding/json"
	"io"
	"io/fs"
	"iter"
	"math/rand"
)

func NewRandomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// loadJSONFromFile loads JSON data from a file and unmarshals it into the provided destination structure.
func loadJSONFromFile(file fs.File, dest interface{}) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return err
	}

	return nil
}

func SeqApply[T any, S any](seq iter.Seq[T], f func(T) S) iter.Seq[S] {
	return func(yield func(S) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

func SliceToPointerSlice[T any](slice []T) []*T {
	result := make([]*T, len(slice))
	for i := range slice {
		result[i] = &slice[i]
	}
	return result
}

func SeqToPointer[T any](seq iter.Seq[T]) iter.Seq[*T] {
	return SeqApply(seq, func(v T) *T {
		return &v
	})
}

func SeqToValue[T any](seq iter.Seq[*T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if v != nil && !yield(*v) {
				return
			}
		}
	}
}
