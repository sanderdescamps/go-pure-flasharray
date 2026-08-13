package fakearray

import (
	"fmt"
	"slices"

	faclient "github.com/sanderdescamps/go-purefa"
)

func (array *Array) GetHostGroups() []faclient.HostGroup {
	hostGroups := make([]faclient.HostGroup, len(array.HostGroups))
	for i, hg := range array.HostGroups {
		hostGroups[i] = *hg
	}
	return hostGroups
}

func (array *Array) GetHostGroup(name string) (*faclient.HostGroup, error) {
	for _, hg := range array.HostGroups {
		if hg.Name == name {
			return hg, nil
		}
	}
	return nil, fmt.Errorf("host group with name %s not found", name)
}

func NewHostGroup(name string) *faclient.HostGroup {
	return &faclient.HostGroup{
		HostGroupShort: faclient.HostGroupShort{
			NoIdReference: faclient.NoIdReference{
				Name: name,
			},
		},
		HostCount: 0,
		Destroyed: false,
	}
}

func (array *Array) AddHostGroup(hg faclient.HostGroup) (*faclient.HostGroup, error) {
	for _, h := range array.HostGroups {
		if h.Name == hg.Name {
			return nil, fmt.Errorf("host group with name %s already exists", hg.Name)
		}
	}
	array.HostGroups = append(array.HostGroups, &hg)
	return &hg, nil
}

func (array *Array) UpdateHostGroup(name string, hgPatch faclient.HostGroupPatchBody) (*faclient.HostGroup, error) {
	for i, hg := range array.HostGroups {
		if hg.Name == name {
			if hgPatch.Name != "" {
				hg.Name = hgPatch.Name
			}
			array.HostGroups[i] = hg
			return hg, nil
		}
	}
	return nil, fmt.Errorf("host group with name %s not found", name)
}

func (array *Array) DeleteHostGroup(name string) error {
	for i, hg := range array.HostGroups {
		if hg.Name == name {
			array.HostGroups = append(array.HostGroups[:i], array.HostGroups[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("host group with name %s not found", name)
}

// // Returns all the names of the hosts that are members of the specified host group. If the host group does not exist, it returns an error.
// func (array *Array) HostGroupGetMembers(hostGroupName string) ([]string, error) {
// 	_, err := array.GetHostGroup(hostGroupName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	members := []string{}
// 	for _, host := range array.Hosts {
// 		if host.HostGroup.Name == hostGroupName {
// 			members = append(members, host.Name)
// 		}
// 	}
// 	return members, nil
// }

func (array *Array) GetHostGroupMembers(hostGroupNames []string, hostNames []string) ([]faclient.HostGroupMember, error) {
	members := []faclient.HostGroupMember{}
	for _, host := range array.Hosts {
		if len(hostNames) < 1 || slices.Contains(hostNames, host.Name) {
			fmt.Printf("Check if groups %v contains %s\n", hostGroupNames, host.HostGroup.Name)
			if len(hostGroupNames) < 1 || slices.Contains(hostGroupNames, host.HostGroup.Name) {
				members = append(members, faclient.HostGroupMember{
					Group:  host.HostGroup,
					Member: host.HostShort,
				})
			}
		}
	}
	return members, nil
}

func (array *Array) AddHostGroupMembers(hostGroupName string, hostNames []string) ([]faclient.HostGroupMember, error) {
	group, err := array.GetHostGroup(hostGroupName)
	if err != nil {
		return nil, err
	}

	members := []faclient.HostGroupMember{}
	for _, host := range array.Hosts {
		if slices.Contains(hostNames, host.Name) {
			array.UpdateHost(host.Name, faclient.HostPatchBody{
				HostGroup: &group.HostGroupShort,
			})
			members = append(members, faclient.HostGroupMember{
				Group:  group.HostGroupShort,
				Member: host.HostShort,
			})
		}
	}
	return members, nil
}

func (array *Array) DeleteHostGroupMembers(hostGroupName string, hostNames []string) error {
	for _, host := range array.Hosts {
		if slices.Contains(hostNames, host.Name) {
			host.HostGroup = faclient.HostGroupShort{}
		}
	}
	return nil
}
