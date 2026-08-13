package flashclient

import (
	"fmt"
	"net/http"
)

type Hardware struct {
	NoIdReference
	Details         string `json:"details"`
	IdentityEnabled bool   `json:"identity_enabled"`
	Index           int    `json:"index"`
	Model           string `json:"model"`
	Serial          string `json:"serial"`
	Slot            int    `json:"slot"`
	Speed           int    `json:"speed"`
	Status          string `json:"status"`
	Temperature     int    `json:"temperature"`
	Type            string `json:"type"`
	Voltage         int    `json:"voltage"`
}

func (fa *FAClient) GetHardware() ([]Hardware, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/hardware"
	result := Results[Hardware]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get hardware, got status code %d", res.StatusCode())
	}
	return result.Items, nil
}
