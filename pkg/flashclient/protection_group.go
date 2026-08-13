package flashclient

import (
	"fmt"
	"net/http"
)

// ProtectionGroup represents a protection group element with properties for
// data protection, replication, and retention management.
type ProtectionGroup struct {
	ProtectionGroupShort

	Context Context `json:"context,omitempty"`

	// Destroyed indicates if the protection group has been destroyed.
	Destroyed bool `json:"destroyed,omitempty"`

	// EradicationConfig contains the eradication settings for the protection group.
	EradicationConfig *ProtectionGroupEradicationConfig `json:"eradication_config,omitempty"`

	// HostCount is the number of hosts in this protection group (read-only).
	HostCount int64 `json:"host_count,omitempty"`

	// HostGroupCount is the number of host groups in this protection group (read-only).
	HostGroupCount int64 `json:"host_group_count,omitempty"`

	// IsLocal indicates if the protection group belongs to the local array (true)
	// or remote array (false). Read-only.
	IsLocal *bool `json:"is_local,omitempty"`

	// Pod references the pod in which the protection group resides.
	Pod *PodShort `json:"pod,omitempty"`

	// ReplicationSchedule contains the schedule settings for asynchronous replication.
	ReplicationSchedule *ReplicationSchedule `json:"replication_schedule,omitempty"`

	// RetentionLock specifies SafeMode restrictions: "ratcheted" or "unlocked".
	// Defaults to "unlocked" for newly created protection groups.
	RetentionLock *string `json:"retention_lock,omitempty"`

	// SnapshotSchedule contains the schedule settings for protection group snapshots.
	SnapshotSchedule *SnapshotSchedule `json:"snapshot_schedule,omitempty"`

	// Source references the array or pod on which the protection group was created.
	Source *Source `json:"source,omitempty"`

	// SourceRetention is the retention policy for the source array of the protection group.
	SourceRetention *RetentionPolicy `json:"source_retention,omitempty"`

	// Space displays provisioned size and physical storage consumption data
	// for the protection group.
	Space *Space `json:"space,omitempty"`

	// TargetCount is the number of targets to which this protection group replicates (read-only).
	TargetCount int64 `json:"target_count,omitempty"`

	// TargetRetention is the retention policy for the target(s) of the protection group.
	TargetRetention *RetentionPolicy `json:"target_retention,omitempty"`

	// TimeRemaining is the time until the destroyed protection group is permanently
	// eradicated, measured in milliseconds (read-only).
	// Before this period elapses, the protection group can be recovered.
	TimeRemaining *int64 `json:"time_remaining,omitempty"`

	// VolumeCount is the number of volumes in the protection group (read-only).
	VolumeCount int64 `json:"volume_count,omitempty"`
}

type ProtectionGroupPatchBody struct {
	Name                *string                           `json:"name,omitempty"`
	Destroyed           *bool                             `json:"destroyed,omitempty"`
	EradicationConfig   *ProtectionGroupEradicationConfig `json:"eradication_config,omitempty"`
	Pod                 *PodShort                         `json:"pod,omitempty"`
	ReplicationSchedule *ReplicationSchedule              `json:"replication_schedule,omitempty"`
	RetentionLock       *string                           `json:"retention_lock,omitempty"`
	SnapshotSchedule    *SnapshotSchedule                 `json:"snapshot_schedule,omitempty"`
	Source              *Source                           `json:"source,omitempty"`
	SourceRetention     *RetentionPolicy                  `json:"source_retention,omitempty"`
	TargetRetention     *RetentionPolicy                  `json:"target_retention,omitempty"`
}

type ProtectionGroupShort struct {
	FixedReference
}

// Placeholder types for referenced schemas
type ProtectionGroupEradicationConfig struct {
	ManualEradication string `json:"manual_eradication,omitempty"`
}

type ReplicationSchedule struct {
	SnapshotSchedule
	// The range of time when to suspend replication.
	// To clear the blackout period, set to an empty string ("").
	Blackout TimeWindow `json:"blackout"`
}

type TimeWindow struct {
	//The window start time. Measured in milliseconds since midnight. The time must be set on the hour. (e.g., `18000000`, which is equal to 5:00 AM).
	Start string `json:"start,omitempty"`
	// The window end time. Measured in milliseconds since midnight. The time must be set on the hour. (e.g., `28800000`, which is equal to 8:00 AM).
	End string `json:"end,omitempty"`
}

type SnapshotSchedule struct {
	// The time of day the snapshot is scheduled to be taken and retained on the local array or immediately replicated to the target(s).
	// Measured in seconds since midnight.
	// The `at` value is only used if the `frequency` parameter is in days
	// (e.g., `259200000`, which is equal to 3 days).
	At int64 `json:"at"`
	// If set to `true`, the policy is enabled.
	Enabled bool `json:"enabled"`
	// The frequency of the scheduled action. Measured in milliseconds.
	Frequency int64 `json:"frequency"`
}

type RetentionPolicy struct {
	// The length of time to keep the specified snapshots. Measured in seconds.
	// Prior to 6.8.2 the range of 60 to 34560000 is accepted.
	// In 6.8.2 and onwards the range of 60 to 2147483647 is accepted.
	AllForSec int32 `json:"all_for_sec"`
	//The number of days to keep the snapshots after the `all_for_sec` period has passed.
	// Prior to 6.6.4 the range of 0 to 4000 is accepted.
	// In 6.6.4 and onwards the range of 0 to 2147483647 is accepted.
	Days int32 `json:"days"`
	// The number of snapshots to keep per day after the `all_for_sec` period has passed.
	// Prior to 6.8.2 the range of 0 to 1440 is accepted.
	// In 6.8.2 and onwards the range of 0 to 2147483647 is accepted.
	PerDay int32 `json:"per_day"`
}

type ProtectionGroupHostMember struct {
	Context Context              `json:"context,omitempty"`
	Group   ProtectionGroupShort `json:"group"`
	Member  HostShort            `json:"member"`
}

type ProtectionGroupHostGroupMember struct {
	Context Context              `json:"context,omitempty"`
	Group   ProtectionGroupShort `json:"group"`
	Member  HostGroupShort       `json:"member"`
}

type ProtectionGroupVolumeMember struct {
	Context Context              `json:"context,omitempty"`
	Group   ProtectionGroupShort `json:"group"`
	Member  VolumeShort          `json:"member"`
}

func (fa *FAClient) GetProtectionGroups() ([]ProtectionGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups"
	result := Results[ProtectionGroup]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection groups, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetProtectionGroup(id string) (*ProtectionGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups"
	result := Results[ProtectionGroup]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("ids", id).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group by id, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("protection group with ID %s not found", id)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) GetProtectionGroupByName(name string) (*ProtectionGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups"
	result := Results[ProtectionGroup]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("names", name).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group by name, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("protection group with name %s not found", name)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) CreateProtectionGroup(name string) (*ProtectionGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups"
	result := Results[ProtectionGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create protection group, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to create protection group, no result returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) UpdateProtectionGroup(id string, patch ProtectionGroupPatchBody) (*ProtectionGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups"
	result := Results[ProtectionGroup]{}
	res, err := fa.RestClient.R().
		SetBody(patch).
		SetQueryParam("ids", id).
		SetResult(&result).
		Patch(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update protection group, got status code %d", res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to update protection group, no result returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) DestroyProtectionGroup(id string) error {
	_, err := fa.UpdateProtectionGroup(id, ProtectionGroupPatchBody{Destroyed: toPtr(true)})
	if err != nil {
		return fmt.Errorf("failed to destroy protection group: %v", err)
	}

	return nil
}

func (fa *FAClient) EradicateProtectionGroup(id string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups"
	res, err := fa.RestClient.R().
		SetQueryParam("ids", id).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to eradicate protection group, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) GetAllProtectionGroupHostMembers() ([]ProtectionGroupHostMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/hosts"
	result := Results[ProtectionGroupHostMember]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group host members, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetProtectionGroupHostMembers(pgId string) ([]ProtectionGroupHostMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/hosts"
	result := Results[ProtectionGroupHostMember]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("group_ids", pgId).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group host members, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) AddHostToProtectionGroup(pgId string, hostName string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/hosts"
	res, err := fa.RestClient.R().
		SetQueryParam("group_ids", pgId).
		SetQueryParam("member_names", hostName).
		Post(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add host to protection group, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) RemoveHostFromProtectionGroup(pgId string, hostName string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/hosts"
	res, err := fa.RestClient.R().
		SetQueryParam("group_ids", pgId).
		SetQueryParam("member_names", hostName).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to remove host from protection group, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) GetAllProtectionGroupHostGroupMembers() ([]ProtectionGroupHostGroupMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/host-groups"
	result := Results[ProtectionGroupHostGroupMember]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group host group members, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetProtectionGroupHostGroupMembers(pgId string) ([]ProtectionGroupHostGroupMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/host-groups"
	result := Results[ProtectionGroupHostGroupMember]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("group_ids", pgId).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group host group members, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) AddHostGroupToProtectionGroup(pgId string, hostGroupName string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/host-groups"
	res, err := fa.RestClient.R().
		SetQueryParam("group_ids", pgId).
		SetQueryParam("member_names", hostGroupName).
		Post(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add host group to protection group, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) RemoveHostGroupFromProtectionGroup(pgId string, hostGroupName string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/host-groups"
	res, err := fa.RestClient.R().
		SetQueryParam("group_ids", pgId).
		SetQueryParam("member_names", hostGroupName).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to remove host group from protection group, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) GetAllProtectionGroupVolumeMembers() ([]ProtectionGroupVolumeMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/volumes"
	result := Results[ProtectionGroupVolumeMember]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group volume members, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetProtectionGroupVolumeMembers(pgId string) ([]ProtectionGroupVolumeMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/volumes"
	result := Results[ProtectionGroupVolumeMember]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("group_ids", pgId).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group volume members, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) AddVolumeToProtectionGroup(pgId string, volumeId string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/volumes"
	res, err := fa.RestClient.R().
		SetQueryParam("group_ids", pgId).
		SetQueryParam("member_ids", volumeId).
		Post(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add volume to protection group, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) RemoveVolumeFromProtectionGroup(pgId string, volumeId string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-groups/volumes"
	res, err := fa.RestClient.R().
		SetQueryParam("group_ids", pgId).
		SetQueryParam("member_ids", volumeId).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to remove volume from protection group, got status code %d", res.StatusCode())
	}

	return nil
}
