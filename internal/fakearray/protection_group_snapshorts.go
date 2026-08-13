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

func ProtectionGroupSnapshotsWithSourceIds(sourceIds ...string) func(*flashclient.ProtectionGroupSnapshot) bool {
	return func(s *flashclient.ProtectionGroupSnapshot) bool {
		return slices.Contains(sourceIds, s.Source.Id)
	}
}

func ProtectionGroupSnapshotsWithSourceNames(names ...string) func(*flashclient.ProtectionGroupSnapshot) bool {
	return func(s *flashclient.ProtectionGroupSnapshot) bool {
		return slices.Contains(names, s.Source.Name)
	}
}

func (array *Array) GetProtectionGroupSnapshot(id string) (*flashclient.ProtectionGroupSnapshot, error) {
	for i := range array.ProtectionGroupSnapshots {
		if array.ProtectionGroupSnapshots[i].Id == id {
			return array.ProtectionGroupSnapshots[i], nil
		}
	}
	return nil, fmt.Errorf("protection group snapshot with ID %s not found", id)
}

// GetProtectionGroupSnapshotByName returns the protection group snapshot with the given name.
func (array *Array) GetProtectionGroupSnapshotByName(name string) (*flashclient.ProtectionGroupSnapshot, error) {
	for i := range array.ProtectionGroupSnapshots {
		if array.ProtectionGroupSnapshots[i].Name == name {
			return array.ProtectionGroupSnapshots[i], nil
		}
	}
	return nil, fmt.Errorf("protection group snapshot with name %s not found", name)
}

// GetProtectionGroupSnapshotsForSources returns all protection group snapshots that match the given source protection group IDs.
func (array *Array) GetProtectionGroupSnapshotsForSources(sourceIds ...string) ([]flashclient.ProtectionGroupSnapshot, error) {
	snaps := []flashclient.ProtectionGroupSnapshot{}
	for i := range array.ProtectionGroupSnapshots {
		if slices.Contains(sourceIds, array.ProtectionGroupSnapshots[i].Source.Id) {
			snaps = append(snaps, *array.ProtectionGroupSnapshots[i])
		}
	}
	return snaps, nil
}

func (array *Array) GetProtectionGroupSnapshots(filters ...func(*flashclient.ProtectionGroupSnapshot) bool) ([]flashclient.ProtectionGroupSnapshot, error) {
	result := []flashclient.ProtectionGroupSnapshot{}
	for _, snapshot := range array.ProtectionGroupSnapshots {
		include := true
		for _, filter := range filters {
			if !filter(snapshot) {
				include = false
				break
			}
		}
		if include {
			result = append(result, *snapshot)
		}
	}
	return result, nil
}

func (array *Array) AddProtectionGroupSnapshot(snapshot flashclient.ProtectionGroupSnapshot) (*flashclient.ProtectionGroupSnapshot, error) {
	if _, err := array.GetProtectionGroupSnapshotByName(snapshot.Name); err == nil {
		return nil, fmt.Errorf("protection group snapshot with name %s already exists: %w", snapshot.Name, ErrAlreadyExists)
	}
	if snapshot.Id == "" {
		snapshot.Id = uuid.New().String()
	}
	array.ProtectionGroupSnapshots = append(array.ProtectionGroupSnapshots, &snapshot)
	return &snapshot, nil
}

func (array *Array) CreateProtectionGroupSnapshot(sourceId string, post flashclient.ProtectionGroupSnapshotPostBody) (*flashclient.ProtectionGroupSnapshot, error) {
	existingSnaps, err := array.GetProtectionGroupSnapshotsForSources(sourceId)
	if err != nil {
		return nil, fmt.Errorf("error checking for existing protection group snapshots: %w", err)
	}

	sourcePG, err := array.GetProtectionGroup(sourceId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving source protection group: %w", err)
	}

	suffix := ""
	if post.Suffix != nil {
		suffix = *post.Suffix
	} else {
		existingSnapsIndexes := []int{}
		re := regexp.MustCompile(sourcePG.Name + ".(\\d+)")
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

	snapshotName := fmt.Sprintf("%s.%s", sourcePG.Name, suffix)
	if _, err := array.GetProtectionGroupSnapshotByName(snapshotName); err == nil {
		return nil, fmt.Errorf("protection group snapshot with name %s already exists: %w", snapshotName, ErrAlreadyExists)
	}

	newSnap := flashclient.ProtectionGroupSnapshot{
		FixedReference: flashclient.FixedReference{
			Name: snapshotName,
			Id:   uuid.New().String(),
		},
		Context: sourcePG.Context,

		Source: flashclient.Source{
			FixedReference: sourcePG.FixedReference,
		},
		Created:   time.Now().UnixMilli(),
		Destroyed: false,
		Suffix:    suffix,
	}

	return array.AddProtectionGroupSnapshot(newSnap)
}

func (array *Array) UpdateProtectionGroupSnapshot(id string, snapshotPatch flashclient.ProtectionGroupSnapshotPatchBody) (*flashclient.ProtectionGroupSnapshot, error) {
	for i := range array.ProtectionGroupSnapshots {
		if array.ProtectionGroupSnapshots[i].Id == id {
			if snapshotPatch.Name != nil {
				array.ProtectionGroupSnapshots[i].Name = *snapshotPatch.Name
			}
			if snapshotPatch.Destroyed != nil {
				array.ProtectionGroupSnapshots[i].Destroyed = *snapshotPatch.Destroyed
			}

			if snapshotPatch.EradicationConfig != nil {
				array.ProtectionGroupSnapshots[i].EradicationConfig = *snapshotPatch.EradicationConfig
			}

			return array.ProtectionGroupSnapshots[i], nil
		}
	}

	return nil, fmt.Errorf("protection group snapshot with ID %s not found", id)
}

func (array *Array) DestroyProtectionGroupSnapshot(id string) error {
	_, err := array.UpdateProtectionGroupSnapshot(id, flashclient.ProtectionGroupSnapshotPatchBody{Destroyed: toPtr(true)})
	return err
}

func (array *Array) EradicateProtectionGroupSnapshot(id string) error {
	for i := range array.ProtectionGroupSnapshots {
		if array.ProtectionGroupSnapshots[i].Id == id {
			if !array.ProtectionGroupSnapshots[i].Destroyed {
				return fmt.Errorf("protection group snapshot with ID %s is not yet destroyed", id)
			}
			array.ProtectionGroupSnapshots = append(array.ProtectionGroupSnapshots[:i], array.ProtectionGroupSnapshots[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("protection group snapshot with ID %s not found", id)
}
