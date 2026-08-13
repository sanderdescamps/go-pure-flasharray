package fakearray

func sliceToPointerSlice[T any](slice []T) []*T {
	result := make([]*T, len(slice))
	for i := range slice {
		result[i] = &slice[i]
	}
	return result
}

func toPtr[T any](value T) *T {
	return &value
}
