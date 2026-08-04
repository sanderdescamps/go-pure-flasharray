package mock

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
	faclient "github.com/sanderdescamps/go-purefa"
)

func (array *Array) GetVolumeGroup(id string) (*faclient.VolumeGroup, error) {
	for i := range array.VolumeGroups {
		if array.VolumeGroups[i].Id == id {
			return array.VolumeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("volume group with ID %s not found", id)
}

func (array *Array) GetVolumeGroupByName(name string) (*faclient.VolumeGroup, error) {
	for i := range array.VolumeGroups {
		if array.VolumeGroups[i].Name == name {
			return array.VolumeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("volume group with name %s not found", name)
}

func (array *Array) GetVolumeGroups() []faclient.VolumeGroup {
	return slices.Collect(SeqToValue(slices.Values(array.VolumeGroups)))
}

func NewVolumeGroup(name string, post *faclient.VolumeGroupPost) *faclient.VolumeGroup {
	priorityAdjustment := faclient.NewPriorityAdjustmentDefault()
	if post != nil && post.PriorityAdjustment != nil {
		priorityAdjustment = *post.PriorityAdjustment
	}

	destroyed := false

	qos := faclient.NewQosDefault()
	if post != nil && post.QoS != nil {
		qos = *post.QoS
	}

	return &faclient.VolumeGroup{
		Id:                 uuid.New().String(),
		Name:               name,
		Destroyed:          destroyed,
		PriorityAdjustment: priorityAdjustment,
		QoS:                qos,
	}
}

func (array *Array) AddVolumeGroup(vg faclient.VolumeGroup) (*faclient.VolumeGroup, error) {
	if _, err := array.GetVolumeGroupByName(vg.Name); err == nil {
		return nil, fmt.Errorf("volume group with name %s already exists: %w", vg.Name, ErrAlreadyExists)
	}
	if vg.Id == "" {
		vg.Id = uuid.New().String()
	}
	array.VolumeGroups = append(array.VolumeGroups, &vg)
	return &vg, nil
}

func (array *Array) UpdateVolumeGroup(id string, vgPatch faclient.VolumeGroupPatch) (*faclient.VolumeGroup, error) {
	for i := range array.VolumeGroups {
		if array.VolumeGroups[i].Id == id {
			if vgPatch.Name != nil {
				array.VolumeGroups[i].Name = *vgPatch.Name
			}
			if vgPatch.Destroyed != nil {
				array.VolumeGroups[i].Destroyed = *vgPatch.Destroyed
			}
			if vgPatch.PriorityAdjustment != nil {
				array.VolumeGroups[i].PriorityAdjustment = *vgPatch.PriorityAdjustment
			}
			if vgPatch.QoS != nil {
				array.VolumeGroups[i].QoS = *vgPatch.QoS
			}
		}
	}

	return nil, fmt.Errorf("volume group with ID %s not found", id)
}

func (array *Array) EradicateVolumeGroup(id string) error {
	for i := range array.VolumeGroups {
		if array.VolumeGroups[i].Id == id {
			array.VolumeGroups = append(array.VolumeGroups[:i], array.VolumeGroups[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("volume group with ID %s not found", id)
}
