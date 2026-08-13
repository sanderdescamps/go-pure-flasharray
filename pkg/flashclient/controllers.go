package flashclient

import (
	"fmt"
	"net/http"
)

type Controller struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	Model     string `json:"model"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	Version   string `json:"version"`
	ModeSince int64  `json:"mode_since"`
}

func (fa *FAClient) GetControllers() ([]Controller, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}
	uri := "/controllers"
	result := Results[Controller]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get controllers, got status code %d", res.StatusCode())
	}
	return result.Items, nil
}
