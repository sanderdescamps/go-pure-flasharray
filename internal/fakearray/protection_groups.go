package fakearray

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

var ErrInvalidProtectionGroup = fmt.Errorf("invalid protection group")

func NewProtectionGroupPost(name string) *flashclient.ProtectionGroup {
	return &flashclient.ProtectionGroup{
		ProtectionGroupShort: flashclient.ProtectionGroupShort{
			FixedReference: flashclient.FixedReference{
				Id:   uuid.New().String(),
				Name: name,
			},
		},
	}
}

func (array *Array) GetProtectionGroup(id string) (*flashclient.ProtectionGroup, error) {
	for _, pg := range array.ProtectionGroups {
		if pg.Id == id {
			return pg, nil
		}
	}
	return nil, fmt.Errorf("protection group with ID %s not found", id)
}

func (array *Array) GetProtectionGroupByName(name string) (*flashclient.ProtectionGroup, error) {
	for _, pg := range array.ProtectionGroups {
		if pg.Name == name {
			return pg, nil
		}
	}
	return nil, fmt.Errorf("protection group with name %s not found", name)
}

func (array *Array) GetProtectionGroups() []flashclient.ProtectionGroup {
	protectionGroups := make([]flashclient.ProtectionGroup, len(array.ProtectionGroups))
	for i, pg := range array.ProtectionGroups {
		protectionGroups[i] = *pg
	}
	return protectionGroups
}

func (array *Array) AddProtectionGroup(pg flashclient.ProtectionGroup) (*flashclient.ProtectionGroup, error) {
	if _, err := array.GetProtectionGroupByName(pg.Name); err == nil {
		return nil, fmt.Errorf("protection group with name %s already exists: %w", pg.Name, ErrAlreadyExists)
	}
	if pg.Id == "" {
		pg.Id = uuid.New().String()
	}
	array.ProtectionGroups = append(array.ProtectionGroups, &pg)
	return &pg, nil
}

func (array *Array) UpdateProtectionGroup(id string, pgPatch flashclient.ProtectionGroupPatchBody) (*flashclient.ProtectionGroup, error) {
	for i := range array.ProtectionGroups {
		if array.ProtectionGroups[i].Id == id {
			if pgPatch.Name != nil {
				array.ProtectionGroups[i].Name = *pgPatch.Name
			}
			if pgPatch.Destroyed != nil {
				array.ProtectionGroups[i].Destroyed = *pgPatch.Destroyed
			}
			if pgPatch.EradicationConfig != nil {
				array.ProtectionGroups[i].EradicationConfig = pgPatch.EradicationConfig
			}
			if pgPatch.Pod != nil {
				array.ProtectionGroups[i].Pod = pgPatch.Pod
			}
			if pgPatch.ReplicationSchedule != nil {
				array.ProtectionGroups[i].ReplicationSchedule = pgPatch.ReplicationSchedule
			}
			if pgPatch.RetentionLock != nil {
				array.ProtectionGroups[i].RetentionLock = pgPatch.RetentionLock
			}
			if pgPatch.SnapshotSchedule != nil {
				array.ProtectionGroups[i].SnapshotSchedule = pgPatch.SnapshotSchedule
			}
			if pgPatch.Source != nil {
				array.ProtectionGroups[i].Source = pgPatch.Source
			}
			if pgPatch.SourceRetention != nil {
				array.ProtectionGroups[i].SourceRetention = pgPatch.SourceRetention
			}
			if pgPatch.TargetRetention != nil {
				array.ProtectionGroups[i].TargetRetention = pgPatch.TargetRetention
			}
			return array.ProtectionGroups[i], nil
		}
	}
	return nil, fmt.Errorf("protection group with ID %s not found", id)
}

func (array *Array) EradicateProtectionGroup(id string) error {
	for i := range array.ProtectionGroups {
		if array.ProtectionGroups[i].Id == id {
			array.ProtectionGroups = append(array.ProtectionGroups[:i], array.ProtectionGroups[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("protection group with ID %s not found", id)
}

// Returns true if the protection group has any host members, false otherwise. Does not check for conflicts with other member types (host group, volume).
func (array *Array) protectionGroupHasHostMembers(pgId string) bool {
	for _, member := range array.ProtectionGroupsHostMembers {
		if member.Group.Id == pgId {
			return true
		}
	}
	return false
}

// Returns true if the protection group has any host group members, false otherwise. Does not check for conflicts with other member types (host, volume).
func (array *Array) protectionGroupHasHostGroupMembers(pgId string) bool {
	for _, member := range array.ProtectionGroupsHostGroupMembers {
		if member.Group.Id == pgId {
			return true
		}
	}
	return false
}

// Returns true if the protection group has any volume members, false otherwise. Does not check for conflicts with other member types (host, host group).
func (array *Array) protectionGroupHasVolumeMembers(pgId string) bool {
	for _, member := range array.ProtectionGroupsVolumeMembers {
		if member.Group.Id == pgId {
			return true
		}
	}
	return false
}

func (array *Array) protectionGroupIsHostMember(pgId string, hostName string) bool {
	for _, member := range array.ProtectionGroupsHostMembers {
		if member.Group.Id == pgId && member.Member.Name == hostName {
			return true
		}
	}
	return false
}

func (array *Array) protectionGroupIsHostGroupMember(pgId string, hostGroupName string) bool {
	for _, member := range array.ProtectionGroupsHostGroupMembers {
		if member.Group.Id == pgId && member.Member.Name == hostGroupName {
			return true
		}
	}
	return false
}

func (array *Array) protectionGroupIsVolumeMember(pgId string, volumeName string) bool {
	for _, member := range array.ProtectionGroupsVolumeMembers {
		if member.Group.Id == pgId && member.Member.Name == volumeName {
			return true
		}
	}
	return false
}

// GetProtectionGroupHostMembers returns members for all protection group ids.
// When pgIds is empty, members for all protection groups are returned. When memberNames is empty, all members are returned.
func (array *Array) GetProtectionGroupHostMembers(pgIds []string, memberNames []string) ([]flashclient.ProtectionGroupHostMember, error) {
	hostMembers := []flashclient.ProtectionGroupHostMember{}
	for _, member := range array.ProtectionGroupsHostMembers {
		if (len(pgIds) == 0 || slices.Contains(pgIds, member.Group.Id)) && (len(memberNames) == 0 || slices.Contains(memberNames, member.Member.Name)) {
			hostMembers = append(hostMembers, *member)
		}
	}

	return hostMembers, nil
}

// GetProtectionGroupHostGroupMembers returns members for all protection group ids.
// When pgIds is empty, members for all protection groups are returned. When memberNames is empty, all members are returned.
func (array *Array) GetProtectionGroupHostGroupMembers(pgIds []string, memberNames []string) ([]flashclient.ProtectionGroupHostGroupMember, error) {
	hostGroupMembers := []flashclient.ProtectionGroupHostGroupMember{}
	for _, member := range array.ProtectionGroupsHostGroupMembers {
		if (len(pgIds) == 0 || slices.Contains(pgIds, member.Group.Id)) && (len(memberNames) == 0 || slices.Contains(memberNames, member.Member.Name)) {
			hostGroupMembers = append(hostGroupMembers, *member)
		}
	}

	return hostGroupMembers, nil
}

// GetProtectionGroupVolumeMembers returns members for all protection group ids.
// When pgIds is empty, members for all protection groups are returned. When memberNames is empty, all members are returned.
func (array *Array) GetProtectionGroupVolumeMembers(pgIds []string, memberIds []string) ([]flashclient.ProtectionGroupVolumeMember, error) {
	volumeMembers := []flashclient.ProtectionGroupVolumeMember{}
	for _, member := range array.ProtectionGroupsVolumeMembers {
		if (len(pgIds) == 0 || slices.Contains(pgIds, member.Group.Id)) && (len(memberIds) == 0 || slices.Contains(memberIds, member.Member.Id)) {
			volumeMembers = append(volumeMembers, *member)
		}
	}

	return volumeMembers, nil
}

// GetHostMembersOfProtectionGroup returns certain host members of a protection group.
// When memberNames is empty, all members are returned. Returns nil if the protection group has no members.
// Returns an error if the protection group has conflicting member types (host group or volume).
func (array *Array) GetHostMembersOfProtectionGroup(pgId string, memberNames []string) ([]flashclient.ProtectionGroupHostMember, error) {
	hasHostMembers := array.protectionGroupHasHostMembers(pgId)
	hasHostGroupMembers := array.protectionGroupHasHostGroupMembers(pgId)
	hasVolumeMembers := array.protectionGroupHasVolumeMembers(pgId)
	if !hasHostMembers && !hasHostGroupMembers && !hasVolumeMembers {
		return nil, nil
	} else if hasHostGroupMembers || hasVolumeMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: %w", ErrInvalidProtectionGroup)
	}
	hostMembers := []flashclient.ProtectionGroupHostMember{}
	for _, member := range array.ProtectionGroupsHostMembers {
		if member.Group.Id == pgId && (len(memberNames) == 0 || slices.Contains(memberNames, member.Member.Name)) {
			hostMembers = append(hostMembers, *member)
		}
	}
	return hostMembers, nil
}

// GetHostGroupMembersOfProtectionGroup returns certain host group members of a protection group.
// When memberNames is empty, all members are returned. Returns nil if the protection group has no members.
// Returns an error if the protection group has conflicting member types (host or volume).
func (array *Array) GetHostGroupMembersOfProtectionGroup(pgId string, memberNames []string) ([]flashclient.ProtectionGroupHostGroupMember, error) {
	hasHostMembers := array.protectionGroupHasHostMembers(pgId)
	hasHostGroupMembers := array.protectionGroupHasHostGroupMembers(pgId)
	hasVolumeMembers := array.protectionGroupHasVolumeMembers(pgId)
	if !hasHostMembers && !hasHostGroupMembers && !hasVolumeMembers {
		return nil, nil
	} else if hasHostMembers || hasVolumeMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: %w", ErrInvalidProtectionGroup)
	}
	hostGroupMembers := []flashclient.ProtectionGroupHostGroupMember{}
	for _, member := range array.ProtectionGroupsHostGroupMembers {
		if member.Group.Id == pgId && (len(memberNames) == 0 || slices.Contains(memberNames, member.Member.Name)) {
			hostGroupMembers = append(hostGroupMembers, *member)
		}
	}
	return hostGroupMembers, nil
}

// GetVolumeMembersOfProtectionGroup returns certain volume members of a protection group.
// When memberNames is empty, all members are returned. Returns nil if the protection group has no members.
// Returns an error if the protection group has conflicting member types (host or host group).
func (array *Array) GetVolumeMembersOfProtectionGroup(pgId string, memberIds []string) ([]flashclient.ProtectionGroupVolumeMember, error) {
	hasHostMembers := array.protectionGroupHasHostMembers(pgId)
	hasHostGroupMembers := array.protectionGroupHasHostGroupMembers(pgId)
	hasVolumeMembers := array.protectionGroupHasVolumeMembers(pgId)
	if !hasHostMembers && !hasHostGroupMembers && !hasVolumeMembers {
		return nil, nil
	} else if hasHostMembers || hasHostGroupMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: %w", ErrInvalidProtectionGroup)
	}
	volumeMembers := []flashclient.ProtectionGroupVolumeMember{}
	for _, member := range array.ProtectionGroupsVolumeMembers {
		if member.Group.Id == pgId && (len(memberIds) == 0 || slices.Contains(memberIds, member.Member.Id)) {
			volumeMembers = append(volumeMembers, *member)
		}
	}
	return volumeMembers, nil
}

func (array *Array) AddProtectionGroupHostMember(pgId string, hostName string) (*flashclient.ProtectionGroupHostMember, error) {
	hasHostGroupMembers := array.protectionGroupHasHostGroupMembers(pgId)
	hasVolumeMembers := array.protectionGroupHasVolumeMembers(pgId)
	if hasHostGroupMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: protection group already has host group members: %w", ErrInvalidProtectionGroup)
	} else if hasVolumeMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: protection group already has volume members: %w", ErrInvalidProtectionGroup)
	}

	pg, err := array.GetProtectionGroup(pgId)
	if err != nil {
		return nil, err
	}

	host, err := array.GetHost(hostName)
	if err != nil {
		return nil, err
	}

	if array.protectionGroupIsHostMember(pg.Id, host.Name) {
		return nil, fmt.Errorf("host %s is already a member of protection group %s: %w", host.Name, pg.Name, ErrAlreadyExists)
	}

	member := &flashclient.ProtectionGroupHostMember{
		Context: pg.Context,
		Group:   pg.ProtectionGroupShort,
		Member:  host.HostShort,
	}

	array.ProtectionGroupsHostMembers = append(array.ProtectionGroupsHostMembers, member)
	return member, nil
}

func (array *Array) AddProtectionGroupHostGroupMember(pgId string, hostGroupName string) (*flashclient.ProtectionGroupHostGroupMember, error) {
	hasHostMembers := array.protectionGroupHasHostMembers(pgId)
	hasVolumeMembers := array.protectionGroupHasVolumeMembers(pgId)
	if hasHostMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: protection group already has host members: %w", ErrInvalidProtectionGroup)
	} else if hasVolumeMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: protection group already has volume members: %w", ErrInvalidProtectionGroup)
	}

	pg, err := array.GetProtectionGroup(pgId)
	if err != nil {
		return nil, err
	}

	hostGroup, err := array.GetHostGroup(hostGroupName)
	if err != nil {
		return nil, err
	}

	if array.protectionGroupIsHostGroupMember(pg.Id, hostGroup.Name) {
		return nil, fmt.Errorf("host group %s is already a member of protection group %s: %w", hostGroup.Name, pg.Name, ErrAlreadyExists)
	}

	member := &flashclient.ProtectionGroupHostGroupMember{
		Context: pg.Context,
		Group:   pg.ProtectionGroupShort,
		Member:  hostGroup.HostGroupShort,
	}

	array.ProtectionGroupsHostGroupMembers = append(array.ProtectionGroupsHostGroupMembers, member)
	return member, nil
}

func (array *Array) AddProtectionGroupVolumeMember(pgId string, volumeId string) (*flashclient.ProtectionGroupVolumeMember, error) {
	hasHostMembers := array.protectionGroupHasHostMembers(pgId)
	hasHostGroupMembers := array.protectionGroupHasHostGroupMembers(pgId)
	if hasHostMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: protection group already has host members: %w", ErrInvalidProtectionGroup)
	} else if hasHostGroupMembers {
		return nil, fmt.Errorf("protection group can only have one type of members: protection group already has host group members: %w", ErrInvalidProtectionGroup)
	}

	pg, err := array.GetProtectionGroup(pgId)
	if err != nil {
		return nil, err
	}

	volume, err := array.GetVolume(volumeId)
	if err != nil {
		return nil, err
	}

	if array.protectionGroupIsVolumeMember(pg.Id, volume.Id) {
		return nil, fmt.Errorf("volume %s is already a member of protection group %s: %w", volume.Name, pg.Name, ErrAlreadyExists)
	}

	member := &flashclient.ProtectionGroupVolumeMember{
		Context: pg.Context,
		Group:   pg.ProtectionGroupShort,
		Member:  volume.VolumeShort,
	}

	array.ProtectionGroupsVolumeMembers = append(array.ProtectionGroupsVolumeMembers, member)
	return member, nil
}

func (array *Array) RemoveProtectionGroupHostMember(pgId string, hostName string) error {
	for i, member := range array.ProtectionGroupsHostMembers {
		if member.Group.Id == pgId && member.Member.Name == hostName {
			array.ProtectionGroupsHostMembers = append(array.ProtectionGroupsHostMembers[:i], array.ProtectionGroupsHostMembers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("host %s is not a member of protection group %s: %w", hostName, pgId, ErrNotFound)
}

func (array *Array) RemoveProtectionGroupHostGroupMember(pgId string, hostGroupName string) error {
	for i, member := range array.ProtectionGroupsHostGroupMembers {
		if member.Group.Id == pgId && member.Member.Name == hostGroupName {
			array.ProtectionGroupsHostGroupMembers = append(array.ProtectionGroupsHostGroupMembers[:i], array.ProtectionGroupsHostGroupMembers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("host group %s is not a member of protection group %s: %w", hostGroupName, pgId, ErrNotFound)
}

func (array *Array) RemoveProtectionGroupVolumeMember(pgId string, volumeId string) error {
	for i, member := range array.ProtectionGroupsVolumeMembers {
		if member.Group.Id == pgId && member.Member.Id == volumeId {
			array.ProtectionGroupsVolumeMembers = append(array.ProtectionGroupsVolumeMembers[:i], array.ProtectionGroupsVolumeMembers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("volume %s is not a member of protection group %s: %w", volumeId, pgId, ErrNotFound)
}
