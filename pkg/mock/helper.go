package mock

import (
	"encoding/json"
	"iter"
	"math/rand"
	"net/http"
	"strings"
)

func NewRandomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
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

func splitQueryParam(param string) []string {
	if param == "" {
		return []string{}
	}
	return strings.Split(param, ",")
}

type ErrorResponse struct {
	Errors []ErrorMsg `json:"errors"`
}

type ErrorMsg struct {
	Context interface{} `json:"context"`
	Message string      `json:"message"`
}

func httpJsonError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	body := ErrorResponse{
		Errors: []ErrorMsg{
			{
				Context: nil,
				Message: msg,
			},
		},
	}
	json.NewEncoder(w).Encode(body)
}
