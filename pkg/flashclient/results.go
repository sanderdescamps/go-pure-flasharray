package flashclient

type Results[T any] struct {
	ContinuationToken  *string `json:"continuation_token,omitempty"`
	Items              []T     `json:"items"`
	TotalItemCount     int     `json:"total_item_count"`
	MoreItemsRemaining bool    `json:"more_items_remaining,omitempty"`
}

func NewResults[T any](items []T) Results[T] {
	return Results[T]{
		Items:              items,
		TotalItemCount:     len(items),
		MoreItemsRemaining: false,
	}
}
