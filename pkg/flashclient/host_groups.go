package flashclient

import (
	"fmt"
	"net/http"
	"strings"
)

type HostGroup struct {
	HostGroupShort

	// ConnectionCount is the number of volumes connected to the host group
	ConnectionCount int64 `json:"connection_count,omitempty"`

	// Destroyed indicates if the host group has been destroyed and is pending eradication
	Destroyed bool `json:"destroyed,omitempty"`

	// HostCount is the number of hosts in the host group
	HostCount int64 `json:"host_count,omitempty"`

	// IsLocal indicates if the host group belongs to the current array (true)
	// or a remote array (false)
	IsLocal bool `json:"is_local,omitempty"`

	// Space displays provisioned and physical storage consumption information
	Space *Space `json:"space,omitempty"`

	// TimeRemaining is the time in milliseconds until the destroyed host group
	// is permanently eradicated
	TimeRemaining int64 `json:"time_remaining,omitempty"`
}

type HostGroupShort struct {
	NoIdReference
}

type HostGroupPatchBody struct {
	// A user-specified name. The name must be locally unique and can be changed.
	Name string `json:"name,omitempty"`
}

type HostGroupMember struct {
	Group  HostGroupShort `json:"group,omitempty"`
	Member HostShort      `json:"member"`
}

func (fa *FAClient) GetHostGroups() ([]HostGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups"
	result := Results[HostGroup]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get host groups, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetHostGroup(name string) (*HostGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups"
	result := Results[HostGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get host group %s, got status code %d", name, res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("host group %s not found", name)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) CreateHostGroup(name string) (*HostGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups"
	result := Results[HostGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create host group %s, got status code %d", name, res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to create host group, no host group returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) UpdateHostGroup(name string, patch HostGroupPatchBody) (*HostGroup, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups"
	result := Results[HostGroup]{}
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		SetBody(&patch).
		SetResult(&result).
		Patch(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update host group %s, got status code %d", name, res.StatusCode())
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("failed to update host group, no host group returned")
	}

	return &result.Items[0], nil
}

func (fa *FAClient) DeleteHostGroup(name string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups"
	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete host group %s, got status code %d", name, res.StatusCode())
	}

	return nil
}

// GetHostGroupMembers returns all the HostGroupsMembers of all the hosts that are member of certain host groups.
// If hostGroupNames is empty, it returns all the members of all the host groups.
// If hostNames is empty, it returns all the members of the specified host groups.
func (fa *FAClient) GetHostGroupMembers(hostGroupName []string, hostNames []string) ([]HostGroupMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups/hosts"
	result := Results[HostGroupMember]{}
	req := fa.RestClient.R().
		SetResult(&result)

	if len(hostGroupName) > 0 {
		req.SetQueryParam("group_names", strings.Join(hostGroupName, ","))
	}
	if len(hostNames) > 0 {
		req.SetQueryParam("member_names", strings.Join(hostNames, ","))
	}

	res, err := req.Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get host group members for %s, got status code %d", hostGroupName, res.StatusCode())
	}

	return result.Items, nil
}

// AddHostGroupMembers adds members to a host group. It returns the added members.
func (fa *FAClient) AddHostGroupMembers(hostGroupName string, hostNames []string) ([]HostGroupMember, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups/hosts"
	result := Results[HostGroupMember]{}
	res, err := fa.RestClient.R().
		SetQueryParam("group_names", hostGroupName).
		SetQueryParam("member_names", strings.Join(hostNames, ",")).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to add members to host group %s, got status code %d", hostGroupName, res.StatusCode())
	}

	return result.Items, nil
}

// DeleteHostGroupMembers removes members from a host group.
func (fa *FAClient) DeleteHostGroupMembers(hostGroupName string, hostNames []string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/host-groups/hosts"
	res, err := fa.RestClient.R().
		SetQueryParam("group_names", hostGroupName).
		SetQueryParam("member_names", strings.Join(hostNames, ",")).
		Delete(uri)
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete members from host group %s, got status code %d", hostGroupName, res.StatusCode())
	}

	return nil
}
