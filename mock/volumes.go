package mock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	faclient "github.com/sanderdescamps/go-purefa"
)

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrAlreadyExists = fmt.Errorf("already exists")
)

func NewVolumeFromVolumePost(volumePost faclient.VolumePost, name string) faclient.Volume {
	priorityAdjustment := faclient.NewPriorityAdjustmentDefault()
	if volumePost.PriorityAdjustment != nil {
		priorityAdjustment = *volumePost.PriorityAdjustment
	}

	destroyed := false
	if volumePost.Destroyed != nil {
		destroyed = *volumePost.Destroyed
	}

	qos := faclient.NewQosDefault()
	if volumePost.QoS != nil {
		qos = *volumePost.QoS
	}

	newVolume := faclient.Volume{
		Id:                      uuid.New().String(),
		Name:                    name,
		ConnectionCount:         0,
		Created:                 time.Now().UnixMilli(),
		Destroyed:               destroyed,
		HostEncryptionKeyStatus: "none",
		PriorityAdjustment:      priorityAdjustment,
		Provisioned:             volumePost.Provisioned,
		QoS:                     qos,
		Serial:                  NewVolumeSerial(),
		Space:                   faclient.Space{},
		TimeRemaining:           0,
		Pod:                     nil,
		Source:                  nil,
		Subtype:                 "regular",
		VolumeGroup:             nil,
		RequestedPromotionState: "",
		PromotionStatus:         "promoted",
		Priority:                0,
	}

	return newVolume
}

func NewVolumeSerial() string {
	const (
		hexChars = "0123456789ABCDEF"
		length   = 24
	)

	b := make([]byte, length)
	for i := range b {
		b[i] = hexChars[rand.Intn(len(hexChars))]
	}
	return string(b)
}

func (array *Array) GetVolume(id string) (*faclient.Volume, error) {
	for i := range array.Volumes {
		if array.Volumes[i].Id == id {
			return array.Volumes[i], nil
		}
	}
	return nil, fmt.Errorf("volume with ID %s not found", id)
}

func (array *Array) GetVolumeByName(name string) (faclient.Volume, error) {
	for i := range array.Volumes {
		if array.Volumes[i].Name == name {
			return *array.Volumes[i], nil
		}
	}
	return faclient.Volume{}, fmt.Errorf("volume with name %s not found", name)
}

func (array *Array) GetVolumes() []faclient.Volume {
	result := make([]faclient.Volume, 0, len(array.Volumes))
	for _, volume := range array.Volumes {
		result = append(result, *volume)
	}
	return result
	// return slices.Collect(maps.Values(array.Volumes)) --- IGNORE ---
}

func (array *Array) AddVolume(volume faclient.Volume) (*faclient.Volume, error) {
	if _, err := array.GetVolumeByName(volume.Name); err == nil {
		return nil, fmt.Errorf("volume with name %s already exists: %w", volume.Name, ErrAlreadyExists)
	}
	if volume.Id == "" {
		volume.Id = uuid.New().String()
	}
	array.Volumes = append(array.Volumes, &volume)
	return &volume, nil
}

func (array *Array) UpdateVolume(id string, volumePatch faclient.VolumePatch) (*faclient.Volume, error) {
	for i := range array.Volumes {
		if array.Volumes[i].Id == id {
			if volumePatch.Name != nil {
				array.Volumes[i].Name = *volumePatch.Name
			}
			if volumePatch.Destroyed != nil {
				array.Volumes[i].Destroyed = *volumePatch.Destroyed
			}
			if volumePatch.Pod != nil {
				array.Volumes[i].Pod = volumePatch.Pod
			}
			if volumePatch.PriorityAdjustment != nil {
				array.Volumes[i].PriorityAdjustment = *volumePatch.PriorityAdjustment
			}
			if volumePatch.Provisioned != nil {
				array.Volumes[i].Provisioned = *volumePatch.Provisioned
			}
			if volumePatch.QoS != nil {
				array.Volumes[i].QoS = *volumePatch.QoS
			}
			if volumePatch.RequestedPromotionState != nil {
				array.Volumes[i].RequestedPromotionState = *volumePatch.RequestedPromotionState
			}
			if volumePatch.VolumeGroup != nil {
				array.Volumes[i].VolumeGroup = volumePatch.VolumeGroup
			}

			return array.Volumes[i], nil
		}
	}

	return nil, fmt.Errorf("volume with ID %s not found", id)

}

func (array *Array) EradicateVolume(id string) error {
	for i := range array.Volumes {
		if array.Volumes[i].Id == id {
			array.Volumes = append(array.Volumes[:i], array.Volumes[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("volume with ID %s not found", id)
}
