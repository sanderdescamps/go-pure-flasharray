package mock

import (
	"fmt"

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
		Name:      name,
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
