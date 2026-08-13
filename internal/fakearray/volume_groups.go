package fakearray

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func (array *Array) GetVolumeGroup(id string) (*flashclient.VolumeGroup, error) {
	for i := range array.VolumeGroups {
		if array.VolumeGroups[i].Id == id {
			return array.VolumeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("volume group with ID %s not found", id)
}

func (array *Array) GetVolumeGroupByName(name string) (*flashclient.VolumeGroup, error) {
	for i := range array.VolumeGroups {
		if array.VolumeGroups[i].Name == name {
			return array.VolumeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("volume group with name %s not found", name)
}

func (array *Array) GetVolumeGroups() []flashclient.VolumeGroup {
	volumeGroups := make([]flashclient.VolumeGroup, len(array.VolumeGroups))
	for i, vg := range array.VolumeGroups {
		volumeGroups[i] = *vg
	}
	return volumeGroups
	// return slices.Collect(SeqToValue(slices.Values(array.VolumeGroups))) --- IGNORE ---
}

func NewVolumeGroup(name string, post *flashclient.VolumeGroupPost) *flashclient.VolumeGroup {
	priorityAdjustment := flashclient.NewPriorityAdjustmentDefault()
	if post != nil && post.PriorityAdjustment != nil {
		priorityAdjustment = *post.PriorityAdjustment
	}

	destroyed := false

	qos := flashclient.NewQosDefault()
	if post != nil && post.QoS != nil {
		qos = *post.QoS
	}

	return &flashclient.VolumeGroup{
		VolumeGroupShort: flashclient.VolumeGroupShort{
			FixedReference: flashclient.FixedReference{
				Id:   uuid.New().String(),
				Name: name,
			},
		},
		Destroyed:          destroyed,
		PriorityAdjustment: priorityAdjustment,
		QoS:                qos,
	}
}

func (array *Array) AddVolumeGroup(vg flashclient.VolumeGroup) (*flashclient.VolumeGroup, error) {
	if _, err := array.GetVolumeGroupByName(vg.Name); err == nil {
		return nil, fmt.Errorf("volume group with name %s already exists: %w", vg.Name, ErrAlreadyExists)
	}
	if vg.Id == "" {
		vg.Id = uuid.New().String()
	}
	array.VolumeGroups = append(array.VolumeGroups, &vg)
	return &vg, nil
}

func (array *Array) CreateVolumeGroup(name string, post flashclient.VolumeGroupPost) (*flashclient.VolumeGroup, error) {
	newVolumeGroup := NewVolumeGroup(name, &post)
	return array.AddVolumeGroup(*newVolumeGroup)
}

func (array *Array) UpdateVolumeGroup(id string, vgPatch flashclient.VolumeGroupPatch) (*flashclient.VolumeGroup, error) {
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
			return array.VolumeGroups[i], nil
		}
	}

	return nil, fmt.Errorf("volume group with ID %s not found", id)
}

func (array *Array) DestroyVolumeGroup(id string) error {
	_, err := array.UpdateVolumeGroup(id, flashclient.VolumeGroupPatch{Destroyed: toPtr(true)})
	return err
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

func (array *Array) GetVolumeGroupMembers(volumeGroupNames []string, volumeNames []string) ([]flashclient.VolumeGroupMember, error) {
	members := []flashclient.VolumeGroupMember{}
	for _, volume := range array.Volumes {
		if len(volumeNames) < 1 || slices.Contains(volumeNames, volume.Name) {
			if len(volumeGroupNames) < 1 || slices.Contains(volumeGroupNames, volume.VolumeGroup.Name) {
				members = append(members, flashclient.VolumeGroupMember{
					Group:  *volume.VolumeGroup,
					Member: volume.VolumeShort,
				})
			}
		}
	}
	return members, nil
}
