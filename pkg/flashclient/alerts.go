package flashclient

import (
	"fmt"
	"net/http"
)

type Alert struct {
	FixedReference
	Actual           string `json:"actual"`
	Category         string `json:"category"`
	Closed           int64  `json:"closed"`
	Code             int64  `json:"code"`
	ComponentName    string `json:"component_name"`
	ComponentType    string `json:"component_type"`
	Created          int64  `json:"created"`
	Description      string `json:"description"`
	Expected         string `json:"expected"`
	Flagged          bool   `json:"flagged"`
	Issue            string `json:"issue"`
	KnowledgeBaseUrl string `json:"knowledge_base_url"`
	Notified         int64  `json:"notified"`
	Severity         string `json:"severity"`
	State            string `json:"state"`
	Summary          string `json:"summary"`
	Updated          int64  `json:"updated"`
}

func (fa *FAClient) GetAlerts() ([]Alert, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/alerts"
	result := Results[Alert]{}
	res, err := fa.RestClient.R().SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get alerts, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}
