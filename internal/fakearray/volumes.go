package fakearray

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
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

// splitVolumeName splits a volume name into its volume group and volume name components.
// If the volume name does not contain a volume group, the first return value will be an empty string.
func splitVolumeName(name string) (string, string) {
	vgroupRegex := regexp.MustCompile(`^([^/]+)/([^/]+)$`)
	if vgroupRegex.MatchString(name) {
		subMatches := vgroupRegex.FindStringSubmatch(name)
		return subMatches[1], subMatches[2]
	}
	return "", name
}

func (array *Array) AddVolume(volume flashclient.Volume) (*flashclient.Volume, error) {
	if _, err := array.GetVolumeByName(volume.Name); err == nil {
		return nil, fmt.Errorf("volume with name %s already exists: %w", volume.Name, ErrAlreadyExists)
	}
	if volume.Id == "" {
		volume.Id = uuid.New().String()
	}

	_, volumeGroupName := splitVolumeName(volume.Name)
	if volumeGroupName != volume.Name {
		volumeGroup, err := array.GetVolumeGroupByName(volumeGroupName)
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("volume group with name %s does not exist: %w", volumeGroupName, ErrNotFound)
		} else if err != nil {
			return nil, fmt.Errorf("failed to get volume group by name %s: %v", volumeGroupName, err)
		}

		volume.VolumeGroup = &volumeGroup.VolumeGroupShort
	}

	array.Volumes = append(array.Volumes, &volume)
	return &volume, nil
}

func (array *Array) UpdateVolume(id string, volumePatch flashclient.VolumePatch) (*flashclient.Volume, error) {
	for i := range array.Volumes {
		if array.Volumes[i].Id == id {
			if volumePatch.Name != nil {
				oldVolumeGroupName, _ := splitVolumeName(array.Volumes[i].Name)
				newVolumeGroupName, _ := splitVolumeName(*volumePatch.Name)
				if oldVolumeGroupName != newVolumeGroupName {
					if newVolumeGroupName == "" {
						// Moving volume out of any volume group
						array.Volumes[i].VolumeGroup = nil
						array.logDebug("Volume with id=%s moved out of any volume group", array.Volumes[i].Id)
					} else {
						volumeGroup, err := array.GetVolumeGroupByName(newVolumeGroupName)
						if errors.Is(err, ErrNotFound) {
							return nil, fmt.Errorf("volume group with name %s does not exist: %w", newVolumeGroupName, ErrNotFound)
						} else if err != nil {
							return nil, fmt.Errorf("failed to get volume group by name %s: %v", newVolumeGroupName, err)
						}
						array.logDebug("Volume with id=%s moved into volume group %s", array.Volumes[i].Id, newVolumeGroupName)
						array.Volumes[i].VolumeGroup = &volumeGroup.VolumeGroupShort
					}
				}
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
