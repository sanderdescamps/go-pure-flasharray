package flashclient

import (
	"fmt"
	"net/http"
)

type VolumeGroup struct {
	VolumeGroupShort
	QoS                Qos                `json:"qos"`
	Destroyed          bool               `json:"destroyed"`
	Pod                PodShort           `json:"pod"`
	PriorityAdjustment PriorityAdjustment `json:"priority_adjustment"`
	Space              Space              `json:"space"`
	// The amount of time left until the destroyed volume group is permanently eradicated, measured in milliseconds.
	TimeRemaining int64 `json:"time_remaining"`
}

type VolumeGroupShort struct {
	FixedReference
}

type VolumeGroupPost struct {
	PriorityAdjustment *PriorityAdjustment `json:"priority_adjustment"`
	QoS                *Qos                `json:"qos"`
}

type VolumeGroupPatch struct {
	Name               *string             `json:"name"`
	Destroyed          *bool               `json:"destroyed"`
	PriorityAdjustment *PriorityAdjustment `json:"priority_adjustment"`
	QoS                *Qos                `json:"qos"`
}

type VolumeGroupMember struct {
	Group  VolumeGroupShort `json:"group"`
	Member VolumeShort      `json:"member"`
}

func (fa *FAClient) GetVolumeGroups() ([]VolumeGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volume-groups"
	result := Results[VolumeGroup]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get volume groups, got status code %d", res.StatusCode())
	}
	return result.Items, nil
}

func (fa *FAClient) GetVolumeGroup(id string) (*VolumeGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volume-groups"
	result := Results[VolumeGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get volume groups, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("volume group with ID %s not found", id)
	}
	return &result.Items[0], nil
}

func (fa *FAClient) GetVolumeGroupByName(name string) (*VolumeGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volume-groups"
	result := Results[VolumeGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get volume groups, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("volume group with name %s not found", name)
	}
	return &result.Items[0], nil
}

func (fa *FAClient) CreateVolumeGroup(name string, vGroupPost VolumeGroupPost) (*VolumeGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volume-groups"
	result := Results[VolumeGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetBody(&vGroupPost).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create volume group, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to create volume group, no volume group returned")
	}
	return &result.Items[0], nil
}

func (fa *FAClient) UpdateVolumeGroup(id string, vGroupPatch VolumeGroupPatch) (*VolumeGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volume-groups"
	result := Results[VolumeGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		SetBody(&vGroupPatch).
		SetResult(&result).
		Patch(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update volume group, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to update volume group, no volume group returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) DestroyVolumeGroup(id string) error {
	_, err := fa.UpdateVolumeGroup(id, VolumeGroupPatch{
		Destroyed: toPtr(true),
	})
	return err
}

func (fa *FAClient) EradicateVolumeGroup(id string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volume-groups"
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to eradicate volume group, got status code %d", res.StatusCode())
	}
	return nil
}
