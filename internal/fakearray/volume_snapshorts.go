package fakearray

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func (array *Array) GetVolumeSnapshot(id string) (*flashclient.VolumeSnapshot, error) {
	for i := range array.VolumeSnapshots {
		if array.VolumeSnapshots[i].Id == id {
			return array.VolumeSnapshots[i], nil
		}
	}
	return nil, fmt.Errorf("volume snapshot with ID %s not found", id)
}

// GetVolumeSnapshotByName returns the volume snapshot with the given name.
func (array *Array) GetVolumeSnapshotByName(name string) (*flashclient.VolumeSnapshot, error) {
	for i := range array.VolumeSnapshots {
		if array.VolumeSnapshots[i].Name == name {
			return array.VolumeSnapshots[i], nil
		}
	}
	return nil, fmt.Errorf("volume snapshot with name %s not found", name)
}

// GetVolumeSnapshotsForSources returns all volume snapshots that match the given source volume names.
func (array *Array) GetVolumeSnapshotsForSources(sourceIds ...string) ([]flashclient.VolumeSnapshot, error) {
	snaps := []flashclient.VolumeSnapshot{}
	for i := range array.VolumeSnapshots {
		if slices.Contains(sourceIds, array.VolumeSnapshots[i].Source.Id) {
			snaps = append(snaps, *array.VolumeSnapshots[i])
		}
	}
	return snaps, nil
}

func (array *Array) GetVolumeSnapshots() []flashclient.VolumeSnapshot {
	result := make([]flashclient.VolumeSnapshot, 0, len(array.VolumeSnapshots))
	for _, snapshot := range array.VolumeSnapshots {
		result = append(result, *snapshot)
	}
	return result
}

func (array *Array) AddVolumeSnapshot(volume flashclient.VolumeSnapshot) (*flashclient.VolumeSnapshot, error) {
	if _, err := array.GetVolumeSnapshotByName(volume.Name); err == nil {
		return nil, fmt.Errorf("volume snapshot with name %s already exists: %w", volume.Name, ErrAlreadyExists)
	}
	if volume.Id == "" {
		volume.Id = uuid.New().String()
	}
	array.VolumeSnapshots = append(array.VolumeSnapshots, &volume)
	return &volume, nil
}

func (array *Array) CreateVolumeSnapshot(sourceId string, post flashclient.VolumeSnapshotPostBody) (*flashclient.VolumeSnapshot, error) {
	existingSnaps, err := array.GetVolumeSnapshotsForSources(sourceId)
	if err != nil {
		return nil, fmt.Errorf("error checking for existing volume snapshots: %w", err)
	}

	sourceVolume, err := array.GetVolume(sourceId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving source volume: %w", err)
	}

	suffix := ""
	if post.Suffix != nil {
		suffix = *post.Suffix
	} else {
		existingSnapsIndexes := []int{}
		re := regexp.MustCompile(sourceVolume.Name + ".(\\d+)")
		for _, snap := range existingSnaps {
			matches := re.FindStringSubmatch(snap.Name)
			if len(matches) > 1 {
				index, _ := strconv.Atoi(matches[1])
				existingSnapsIndexes = append(existingSnapsIndexes, index)
			}
		}
		slices.Sort(existingSnapsIndexes)
		if len(existingSnapsIndexes) == 0 {
			suffix = "001"
		} else {
			suffix = fmt.Sprintf("%03d", existingSnapsIndexes[len(existingSnapsIndexes)-1]+1)
		}
	}

	snapshotName := fmt.Sprintf("%s.%s", sourceVolume.Name, suffix)
	if _, err := array.GetVolumeSnapshotByName(snapshotName); err == nil {
		return nil, fmt.Errorf("volume snapshot with name %s already exists: %w", snapshotName, ErrAlreadyExists)
	}

	newSnap := flashclient.VolumeSnapshot{
		Name: snapshotName,
		Id:   uuid.New().String(),
		Source: flashclient.Source{
			FixedReference: sourceVolume.FixedReference,
		},
		Created:   time.Now().UnixMilli(),
		Destroyed: false,
		Suffix:    suffix,
	}

	if post.Destroyed != nil {
		newSnap.Destroyed = *post.Destroyed
	}

	return array.AddVolumeSnapshot(newSnap)
}

func (array *Array) UpdateVolumeSnapshot(id string, volumePatch flashclient.VolumeSnapshotPatchBody) (*flashclient.VolumeSnapshot, error) {
	for i := range array.VolumeSnapshots {
		if array.VolumeSnapshots[i].Id == id {
			if volumePatch.Name != nil {
				array.VolumeSnapshots[i].Name = *volumePatch.Name
			}
			if volumePatch.Destroyed != nil {
				array.VolumeSnapshots[i].Destroyed = *volumePatch.Destroyed
			}

			return array.VolumeSnapshots[i], nil
		}
	}

	return nil, fmt.Errorf("volume snapshot with ID %s not found", id)
}

func (array *Array) DeleteVolumeSnapshot(id string) error {
	_, err := array.UpdateVolumeSnapshot(id, flashclient.VolumeSnapshotPatchBody{Destroyed: toPtr(true)})
	return err
}

func (array *Array) EradicateVolumeSnapshot(id string) error {
	for i := range array.VolumeSnapshots {
		if array.VolumeSnapshots[i].Id == id {
			if !array.VolumeSnapshots[i].Destroyed {
				return fmt.Errorf("volume snapshot with ID %s is not yet destroyed", id)
			}
			array.VolumeSnapshots = append(array.VolumeSnapshots[:i], array.VolumeSnapshots[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("volume snapshot with ID %s not found", id)
}
