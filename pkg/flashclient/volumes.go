package flashclient

import (
	"fmt"
	"net/http"
)

type Qos struct {
	BandwidthLimit *int64 `json:"bandwidth_limit"`
	IopsLimit      *int64 `json:"iops_limit"`
}

func NewQosDefault() Qos {
	return Qos{
		BandwidthLimit: nil,
		IopsLimit:      nil,
	}
}

func NewPriorityAdjustmentDefault() PriorityAdjustment {
	return PriorityAdjustment{
		PriorityAdjustmentOperator: "+",
		PriorityAdjustmentValue:    0,
	}
}

type PriorityAdjustment struct {
	PriorityAdjustmentOperator string `json:"priority_adjustment_operator"`
	PriorityAdjustmentValue    int32  `json:"priority_adjustment_value"`
}

type Source struct {
	FixedReference
}

type VolumeShort struct {
	FixedReference
}

type Volume struct {
	VolumeShort
	Context         Context `json:"context"`
	ConnectionCount int     `json:"connection_count"`
	// The volume creation time, measured in milliseconds since the UNIX epoch.
	Created int64 `json:"created"`
	// Returns a value of `true` if the volume has been destroyed and is pending eradication.
	Destroyed               bool               `json:"destroyed"`
	HostEncryptionKeyStatus string             `json:"host_encryption_key_status"`
	PriorityAdjustment      PriorityAdjustment `json:"priority_adjustment"`
	Provisioned             int64              `json:"provisioned"`
	QoS                     Qos                `json:"qos"`
	Serial                  string             `json:"serial"`
	Space                   Space              `json:"space"`
	// The amount of time left until the destroyed volume is permanently eradicated, measured in milliseconds.
	// Before the `time_remaining` period has elapsed, the destroyed volume can be recovered by setting `destroyed=false`.
	TimeRemaining           int64             `json:"time_remaining"`
	Pod                     *PodShort         `json:"pod"`
	Source                  *Source           `json:"source"`
	Subtype                 string            `json:"subtype"`
	VolumeGroup             *VolumeGroupShort `json:"volume_group"`
	RequestedPromotionState string            `json:"requested_promotion_state"`
	PromotionStatus         string            `json:"promotion_status"`
	Priority                int               `json:"priority"`
}

type VolumePatch struct {
	Destroyed               *bool               `json:"destroyed,omitempty"`
	Name                    *string             `json:"name,omitempty"`
	Pod                     *PodShort           `json:"pod,omitempty"`
	PriorityAdjustment      *PriorityAdjustment `json:"priority_adjustment,omitempty"`
	Provisioned             *int64              `json:"provisioned,omitempty"`
	QoS                     *Qos                `json:"qos,omitempty"`
	RequestedPromotionState *string             `json:"requested_promotion_state,omitempty"`
	// VolumeGroup             *VolumeGroupShort   `json:"volume_group,omitempty"`
}

type VolumePost struct {
	Destroyed          *bool               `json:"destroyed"`
	PriorityAdjustment *PriorityAdjustment `json:"priority_adjustment"`
	Provisioned        int64               `json:"provisioned"`
	QoS                *Qos                `json:"qos"`
	// The type of volume. Values include `protocol_endpoint` and `regular`.
	Source               Source                 `json:"source"`
	SubType              string                 `json:"subtype"`
	Tags                 []*Tag                 `json:"tags"`
	AddToPromotionGroups []ProtectionGroupShort `json:"add_to_promotion_group"`
}

func (fa *FAClient) GetVolumes() ([]Volume, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volumes"
	result := Results[Volume]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get volumes, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetVolume(id string) (*Volume, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volumes"
	result := Results[Volume]{}
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get volumes, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("volume with ID %s not found", id)
	}
	return &result.Items[0], nil
}

func (fa *FAClient) GetVolumeByName(name string) (*Volume, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volumes"
	result := Results[Volume]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get volumes, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("volume with name %s not found", name)
	}
	return &result.Items[0], nil
}

func (fa *FAClient) UpdateVolume(id string, volumePatch VolumePatch) (*Volume, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volumes"
	result := Results[Volume]{}
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		SetBody(&volumePatch).
		SetResult(&result).
		Patch(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update volume, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to update volume, no volume returned")
	}

	return &result.Items[0], nil
}

// CreateVolume creates a new volume with the specified name and properties.
//
// To add the volume to a volume group, use "{volume-group}/{volume}" as the name, where {volume-group} is the name of
// the volume group and {volume} the name of the new volume. Make sure the volume group already exists before creating the volume.
func (fa *FAClient) CreateVolume(name string, volumePost VolumePost) (*Volume, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volumes"
	result := Results[Volume]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetBody(&volumePost).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create volume, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to create volume, no volume returned")
	}
	return &result.Items[0], nil
}

func (fa *FAClient) EradicateVolume(id string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/volumes"
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to eradicate volume, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) DestroyVolume(id string) error {
	_, err := fa.UpdateVolume(id, VolumePatch{
		Destroyed: new(true),
	})

	return err
}
