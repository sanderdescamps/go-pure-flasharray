package flashclient

import (
	"fmt"
	"net/http"
)

type Drive struct {
	NoIdReference
	Details  string  `json:"details"`
	Capacity float64 `json:"capacity"`
	Protocol string  `json:"protocol"`
	Status   string  `json:"status"`
	Type     string  `json:"type"`
}

func (fa *FAClient) GetDrives() ([]Drive, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/drives"
	result := Results[Drive]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get drives, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}
