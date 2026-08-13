package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func TestVolumeGroups(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("volume_group_list", func(t *testing.T) {
		vgroups, err := client.GetVolumeGroups()
		if err != nil {
			t.Fatalf("Failed to get volume groups: %v", err)
		}
		if lenVGroups := len(vgroups); lenVGroups > 0 {
			t.Logf("%d Volume groups found", lenVGroups)
		} else {
			t.Errorf("Expected at least 1 volume group")
		}
	})

	t.Run("volume_group_create", func(t *testing.T) {
		volumeGroupPost := flashclient.VolumeGroupPost{
			QoS: &flashclient.Qos{
				BandwidthLimit: toPtr[int64](2000),
			},
		}
		name := NewRandomName("volume-group", 8)
		volumeGroup, err := client.CreateVolumeGroup(name, volumeGroupPost)
		if err != nil {
			t.Fatalf("Failed to create volume group: %v", err)
		}
		t.Logf("Volume group created successfully: name=%s, Id=%s", volumeGroup.Name, volumeGroup.Id)

		vgroup, err := client.GetVolumeGroupByName(name)
		if err != nil {
			t.Fatalf("Failed to get volume group by name: %v", err)
		} else if vgroup.Name != name {
			t.Fatalf("Expected volume group name %s, got %s", name, vgroup.Name)
		}
		t.Logf("Volume group retrieved successfully by name: name=%s, Id=%s", vgroup.Name, vgroup.Id)

		vgroup, err = client.GetVolumeGroup(volumeGroup.Id)
		if err != nil {
			t.Fatalf("Failed to get volume group by ID: %v", err)
		} else if vgroup.Name != name {
			t.Fatalf("Expected volume group name %s, got %s", name, vgroup.Name)
		}
		t.Logf("Volume group retrieved successfully by ID: name=%s, Id=%s", vgroup.Name, vgroup.Id)

		err = client.DestroyVolumeGroup(volumeGroup.Id)
		if err != nil {
			t.Fatalf("Failed to destroy volume group by ID: %v", err)
		}
		t.Logf("Volume group destroyed successfully by ID: name=%s, Id=%s", volumeGroup.Name, volumeGroup.Id)

		err = client.EradicateVolumeGroup(volumeGroup.Id)
		if err != nil {
			t.Fatalf("Failed to eradicate volume group by ID: %v", err)
		}
		t.Logf("Volume group eradicated successfully by ID: name=%s, Id=%s", volumeGroup.Name, volumeGroup.Id)
	})

	t.Run("volume_group_update_rename", func(t *testing.T) {
		name := NewRandomName("volume-group", 8)
		vg, err := client.CreateVolumeGroup(name, flashclient.VolumeGroupPost{})
		if err != nil {
			t.Fatalf("Failed to create volume group: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyVolumeGroup(vg.Id)
			client.EradicateVolumeGroup(vg.Id)
		})

		newName := NewRandomName("volume-group", 8)
		updated, err := client.UpdateVolumeGroup(vg.Id, flashclient.VolumeGroupPatch{Name: toPtr(newName)})
		if err != nil {
			t.Fatalf("Failed to rename volume group: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Expected volume group name %s, got %s", newName, updated.Name)
		}
		t.Logf("Volume group renamed: %s -> %s", name, updated.Name)

		retrieved, err := client.GetVolumeGroupByName(newName)
		if err != nil {
			t.Fatalf("Failed to get volume group by new name: %v", err)
		}
		if retrieved.Name != newName {
			t.Errorf("Expected volume group name %s, got %s", newName, retrieved.Name)
		}
	})

	t.Run("volume_group_get_not_found", func(t *testing.T) {
		_, err := client.GetVolumeGroup("00000000-0000-0000-0000-000000000000")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent volume group, got nil")
		}
		t.Logf("Got expected error for non-existent volume group ID: %v", err)
	})

	t.Run("volume_group_get_by_name_not_found", func(t *testing.T) {
		_, err := client.GetVolumeGroupByName("non-existent-volume-group-name")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent volume group by name, got nil")
		}
		t.Logf("Got expected error for non-existent volume group name: %v", err)
	})
}
