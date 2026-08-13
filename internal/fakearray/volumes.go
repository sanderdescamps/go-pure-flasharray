package fakearray

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

const (
	// bytesInKB = 1024 * 1024
	bytesInMB = 1024 * 1024
)

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrAlreadyExists = fmt.Errorf("already exists")
)

func NewVolumeFromPost(name string, volumePost flashclient.VolumePost) flashclient.Volume {
	priorityAdjustment := flashclient.NewPriorityAdjustmentDefault()
	if volumePost.PriorityAdjustment != nil {
		priorityAdjustment = *volumePost.PriorityAdjustment
	}

	destroyed := false
	if volumePost.Destroyed != nil {
		destroyed = *volumePost.Destroyed
	}

	qos := flashclient.NewQosDefault()
	if volumePost.QoS != nil {
		qos = *volumePost.QoS
	}

	provisioned := int64(1) * bytesInMB
	if volumePost.Provisioned != 0 {
		provisioned = volumePost.Provisioned
	}

	newVolume := flashclient.Volume{
		VolumeShort: flashclient.VolumeShort{
			FixedReference: flashclient.FixedReference{
				Id:   uuid.New().String(),
				Name: name,
			},
		},
		ConnectionCount:         0,
		Created:                 time.Now().UnixMilli(),
		Destroyed:               destroyed,
		HostEncryptionKeyStatus: "none",
		PriorityAdjustment:      priorityAdjustment,
		Provisioned:             provisioned,
		QoS:                     qos,
		Serial:                  NewVolumeSerial(),
		Space:                   flashclient.Space{},
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

func (array *Array) GetVolume(id string) (*flashclient.Volume, error) {
	for i := range array.Volumes {
		if array.Volumes[i].Id == id {
			return array.Volumes[i], nil
		}
	}
	return nil, fmt.Errorf("volume with ID %s not found", id)
}

func (array *Array) GetVolumeByName(name string) (flashclient.Volume, error) {
	for i := range array.Volumes {
		if array.Volumes[i].Name == name {
			return *array.Volumes[i], nil
		}
	}
	return flashclient.Volume{}, fmt.Errorf("volume with name %s not found", name)
}

func (array *Array) GetVolumes() []flashclient.Volume {
	result := make([]flashclient.Volume, 0, len(array.Volumes))
	for _, volume := range array.Volumes {
		result = append(result, *volume)
	}
	return result
	// return slices.Collect(maps.Values(array.Volumes)) --- IGNORE ---
}

func (array *Array) AddVolume(volume flashclient.Volume) (*flashclient.Volume, error) {
	if _, err := array.GetVolumeByName(volume.Name); err == nil {
		return nil, fmt.Errorf("volume with name %s already exists: %w", volume.Name, ErrAlreadyExists)
	}
	if volume.Id == "" {
		volume.Id = uuid.New().String()
	}
	array.Volumes = append(array.Volumes, &volume)
	return &volume, nil
}

func (array *Array) UpdateVolume(id string, volumePatch flashclient.VolumePatch) (*flashclient.Volume, error) {
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
