package flashclient

import (
	"fmt"
	"net/http"
)

type Pod struct {
	FixedReference
	Arrays                  []ArrayShort      `json:"arrays"`
	Destroyed               bool              `json:"destroyed"`
	FailoverPreferences     []ArrayTiny       `json:"failover_preferences"`
	Footprint               int               `json:"footprint"`
	Mediator                string            `json:"mediator"`
	MediatorVersion         string            `json:"mediator_version"`
	Source                  Source            `json:"source"`
	Space                   Space             `json:"space"`
	TimeRemaining           int               `json:"time_remaining"`
	RequestedPromotionState string            `json:"requested_promotion_state"`
	PromotionStatus         string            `json:"promotion_status"`
	QuotaLimit              int64             `json:"quota_limit"`
	LinkSourceCount         int               `json:"link_source_count"`
	LinkTargetCount         int               `json:"link_target_count"`
	ArrayCount              int               `json:"array_count"`
	EradicationConfig       EradicationConfig `json:"eradication_config"`
}

type PodPostBody struct {
	FailoverPreferences []ArrayTiny `json:"failover_preferences"`
	QuotaLimit          *int64      `json:"quota_limit"`
	Source              *Source     `json:"source"`
}

type PodPatchBody struct {
	Name                    *string     `json:"name"`
	Destroyed               *bool       `json:"destroyed"`
	FailoverPreferences     []ArrayTiny `json:"failover_preferences"`
	IgnoreUsage             *bool       `json:"ignore_usage"`
	Mediator                *string     `json:"mediator"`
	QuotaLimit              *int64      `json:"quota_limit"`
	Source                  *Source     `json:"source"`
	RequestedPromotionState *string     `json:"requested_promotion_state"`
}

type PodShort struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (fa *FAClient) GetPods() ([]Pod, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/pods"
	result := Results[Pod]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get pods, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetPod(id string) (*Pod, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/pods"
	result := Results[Pod]{}
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get pod with ID %s, got status code %d", id, res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("pod with ID %s not found", id)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) GetPodByName(name string) (*Pod, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/pods"
	result := Results[Pod]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get pod with name %s, got status code %d", name, res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("pod with name %s not found", name)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) CreatePod(name string, podPost PodPostBody) (*Pod, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/pods"
	result := Results[Pod]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetBody(podPost).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create pod with name %s, got status code %d", name, res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to create pod, no pod returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) UpdatePod(id string, podPatch PodPatchBody) (*Pod, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/pods"
	result := Results[Pod]{}
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		SetBody(podPatch).
		SetResult(&result).
		Patch(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update pod with ID %s, got status code %d", id, res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to update pod, no pod returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) DestroyPod(id string) error {
	_, err := fa.UpdatePod(id, PodPatchBody{
		Destroyed: toPtr(true),
	})
	return err
}

func (fa *FAClient) EradicatePod(id string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/pods"
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to eradicate pod with ID %s, got status code %d", id, res.StatusCode())
	}

	return nil
}
