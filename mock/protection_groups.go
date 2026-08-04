package mock

import (
	"fmt"
	"slices"

	faclient "github.com/sanderdescamps/go-purefa"
)

func (array *Array) GetProtectionGroup(id string) (*faclient.ProtectionGroup, error) {
	for _, pg := range array.ProtectionGroups {
		if pg.Id == id {
			return pg, nil
		}
	}
	return nil, fmt.Errorf("protection group with ID %s not found", id)
}

func (array *Array) GetProtectionGroupByName(name string) (*faclient.ProtectionGroup, error) {
	for _, pg := range array.ProtectionGroups {
		if pg.Name == name {
			return pg, nil
		}
	}
	return nil, fmt.Errorf("protection group with name %s not found", name)
}

func (array *Array) GetProtectionGroups() []faclient.ProtectionGroup {
	return slices.Collect(SeqToValue(slices.Values(array.ProtectionGroups)))
}
