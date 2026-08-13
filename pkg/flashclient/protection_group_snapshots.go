package flashclient

import (
	"fmt"
	"net/http"
)

type ProtectionGroupSnapshot struct {
	FixedReference

	// The context in which the operation was performed. Valid values include a reference to any array which is a member of the same fleet. If the array is not a member of a fleet, `context` will always implicitly be set to the array that received the request.  Other parameters provided with the request, such as names of volumes or snapshots, are resolved relative to the provided `context`.
	Context Context `json:"context,omitempty"`

	// The snapshot creation time of the original snapshot source. Measured in milliseconds since the UNIX epoch.
	Created int64 `json:"created,omitempty"`

	// Returns a value of `true` if the protection group snapshot has been destroyed and is pending eradication. The `time_remaining` value displays the amount of time left until the destroyed snapshot is permanently eradicated. Before the `time_remaining` period has elapsed, the destroyed snapshot can be recovered by setting `destroyed=false`. Once the `time_remaining` period has elapsed, the snapshot is permanently eradicated and can no longer be recovered.
	Destroyed bool `json:"destroyed,omitempty"`

	EradicationConfig ProtectionGroupEradicationConfig `json:"eradication_config,omitempty"`

	// The pod in which the protection group of the protection group snapshot resides.
	Pod FixedReference `json:"pod,omitempty"`

	// The original protection group from which this snapshot was taken. For a replicated protection group snapshot being viewed on the target side, the `source` is the replica protection group.
	Source Source `json:"source,omitempty"`

	// Displays provisioned size and physical storage consumption data for each protection group.
	Space Space `json:"space,omitempty"`

	// The name suffix appended to the protection group name to make up the full protection group snapshot name in the form `PGROUP.SUFFIX`. If `suffix` is not specified, the protection group name is in the form `PGROUP.NNN`, where `NNN` is a unique monotonically increasing number. If multiple protection group snapshots are created at a time, the suffix name is appended to those snapshots.
	Suffix string `json:"suffix,omitempty"`

	// The amount of time left until the destroyed snapshot is permanently eradicated. Measured in milliseconds. Before the `time_remaining` period has elapsed, the destroyed snapshot can be recovered by setting `destroyed=false`.
	TimeRemaining int64 `json:"time_remaining,omitempty"`
}

type ProtectionGroupSnapshotPostBody struct {
	EradicationConfig ProtectionGroupEradicationConfig `json:"eradication_config,omitempty"`

	// The name suffix appended to the protection group name to make up the full protection group snapshot name in the form `PGROUP.SUFFIX`. If `suffix` is not specified, the protection group name is in the form `PGROUP.NNN`, where `NNN` is a unique monotonically increasing number. If multiple protection group snapshots are created at a time, the suffix name is appended to those snapshots.
	Suffix *string `json:"suffix,omitempty"`

	// The list of tags to be upserted with the object.
	Tags []Tag `json:"tags,omitempty"`
}

type ProtectionGroupSnapshotPatchBody struct {
	// The new name of the protection group snapshot. The name must be locally unique and can be changed.
	Name *string `json:"name,omitempty"`

	// Returns a value of `true` if the protection group snapshot has been destroyed and is pending eradication. The `time_remaining` value displays the amount of time left until the destroyed snapshot is permanently eradicated. Before the `time_remaining` period has elapsed, the destroyed snapshot can be recovered by setting `destroyed=false`. Once the `time_remaining` period has elapsed, the snapshot is permanently eradicated and can no longer be recovered.
	Destroyed *bool `json:"destroyed,omitempty"`

	EradicationConfig *ProtectionGroupEradicationConfig `json:"eradication_config,omitempty"`
}

func (fa *FAClient) GetProtectionGroupSnapshots() ([]ProtectionGroupSnapshot, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	result := Results[ProtectionGroupSnapshot]{}
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

func (fa *FAClient) GetProtectionGroupSnapshot(id string) (*ProtectionGroupSnapshot, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	result := Results[ProtectionGroupSnapshot]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("ids", id).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group snapshot, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("protection group snapshot %s not found", id)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) GetProtectionGroupSnapshotByName(name string) (*ProtectionGroupSnapshot, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	result := Results[ProtectionGroupSnapshot]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("names", name).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group snapshot, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("protection group snapshot %s not found", name)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) GetProtectionGroupSnapshotsForSource(pgId string) ([]ProtectionGroupSnapshot, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	result := Results[ProtectionGroupSnapshot]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("source_ids", pgId).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get protection group snapshots, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) CreateProtectionGroupSnapshot(pgId string, body ProtectionGroupSnapshotPostBody) (*ProtectionGroupSnapshot, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	result := Results[ProtectionGroupSnapshot]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("source_ids", pgId).
		SetBody(body).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create protection group snapshot, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no protection group snapshot returned after creation")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) UpdateProtectionGroupSnapshot(snapId string, body ProtectionGroupSnapshotPatchBody) (*ProtectionGroupSnapshot, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	result := Results[ProtectionGroupSnapshot]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("ids", snapId).
		SetBody(body).
		Patch(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to patch protection group snapshot, got status code %d", res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no protection group snapshot returned after patch")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) DestroyProtectionGroupSnapshot(snapId string) error {
	_, err := fa.UpdateProtectionGroupSnapshot(snapId, ProtectionGroupSnapshotPatchBody{Destroyed: toPtr(true)})
	if err != nil {
		return fmt.Errorf("failed to destroy protection group snapshot: %v", err)
	}

	return nil
}

func (fa *FAClient) EradicateProtectionGroupSnapshot(snapId string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/protection-group-snapshots"
	res, err := fa.RestClient.R().
		SetQueryParam("ids", snapId).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete protection group snapshot, got status code %d", res.StatusCode())
	}

	return nil
}
