package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
	"github.com/stretchr/testify/assert"
)

func TestVolumeSnapshots(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("volume_snapshot_list", func(t *testing.T) {
		snapshots, err := client.GetVolumeSnapshots()
		if err != nil {
			t.Fatalf("Failed to get volume snapshots: %v", err)
		}
		t.Logf("%d volume snapshots found", len(snapshots))
	})

	t.Run("volume_snapshot_lifecycle", func(t *testing.T) {
		volume, err := client.CreateVolume(NewRandomName("volume", 8), flashclient.VolumePost{
			Provisioned: 1024,
		})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}

		t.Cleanup(func() {
			err := client.DeleteVolume(volume.Id)
			if err != nil {
				t.Fatalf("Failed to delete volume: %v", err)
			}
			err = client.EradicateVolume(volume.Id)
			if err != nil {
				t.Fatalf("Failed to eradicate volume: %v", err)
			}
		})

		snapshots, err := client.GetVolumeSnapshotsForVolume(volume.Id)
		if err != nil {
			t.Fatalf("Failed to get volume snapshots for volume %s: %v", volume.Id, err)
		}
		assert.Equal(t, 0, len(snapshots), "Expected no snapshots for newly created volume")

		snap, err := client.CreateVolumeSnapshot(volume.Id, flashclient.VolumeSnapshotPostBody{})
		if err != nil {
			t.Fatalf("Failed to create volume snapshot for volume %s: %v", volume.Id, err)
		}
		assert.Equal(t, volume.Id, snap.Source.Id, "Snapshot's source ID should match the original volume ID")

		err = client.DestroyVolumeSnapshot(snap.Id)
		if err != nil {
			t.Fatalf("Failed to destroy volume snapshot %s: %v", snap.Id, err)
		}
		err = client.EradicateVolumeSnapshot(snap.Id)
		if err != nil {
			t.Fatalf("Failed to eradicate volume snapshot %s: %v", snap.Id, err)
		}
	})

	t.Run("volume_snapshot_list_for_multiple_volumes", func(t *testing.T) {
		volumes, err := client.GetVolumes()
		if err != nil {
			t.Fatalf("Failed to get volumes: %v", err)
		}
		if len(volumes) < 2 {
			t.Skip("Need at least 2 volumes to test multi-volume snapshot filter")
		}

		volumeIds := []string{volumes[0].Id, volumes[1].Id}
		snapshots, err := client.GetVolumeSnapshotsForVolume(volumeIds...)
		if err != nil {
			t.Fatalf("Failed to get volume snapshots for volumes %v: %v", volumeIds, err)
		}
		t.Logf("%d volume snapshots found for volumes %v", len(snapshots), volumeIds)
	})

	t.Run("volume_snapshot_get_not_found", func(t *testing.T) {
		nonExistentId := "00000000-0000-0000-0000-000000000000"
		_, err := client.GetVolumeSnapshot(nonExistentId)
		if err == nil {
			t.Fatalf("Expected error when getting non-existent volume snapshot, got nil")
		}
		t.Logf("Got expected error for non-existent snapshot: %v", err)
	})

	t.Run("volume_snapshot_get_by_name", func(t *testing.T) {
		volume, err := client.CreateVolume(NewRandomName("volume", 8), flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DeleteVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		snap, err := client.CreateVolumeSnapshot(volume.Id, flashclient.VolumeSnapshotPostBody{})
		if err != nil {
			t.Fatalf("Failed to create volume snapshot: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyVolumeSnapshot(snap.Id)
			client.EradicateVolumeSnapshot(snap.Id)
		})

		retrieved, err := client.GetVolumeSnapshotByName(snap.Name)
		if err != nil {
			t.Fatalf("Failed to get volume snapshot by name %s: %v", snap.Name, err)
		}
		if retrieved.Id != snap.Id {
			t.Errorf("Expected snapshot ID %s, got %s", snap.Id, retrieved.Id)
		}
		t.Logf("Volume snapshot retrieved by name: name=%s, id=%s", retrieved.Name, retrieved.Id)
	})

	t.Run("volume_snapshot_update_rename", func(t *testing.T) {
		volume, err := client.CreateVolume(NewRandomName("volume", 8), flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DeleteVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		snap, err := client.CreateVolumeSnapshot(volume.Id, flashclient.VolumeSnapshotPostBody{})
		if err != nil {
			t.Fatalf("Failed to create volume snapshot: %v", err)
		}

		newSuffix := NewString(6)
		newName := volume.Name + "." + newSuffix
		updated, err := client.UpdateVolumeSnapshot(snap.Id, flashclient.VolumeSnapshotPatchBody{Name: toPtr(newName)})
		if err != nil {
			t.Fatalf("Failed to rename volume snapshot: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Expected snapshot name %s, got %s", newName, updated.Name)
		}
		t.Logf("Volume snapshot renamed: %s -> %s", snap.Name, updated.Name)
		t.Cleanup(func() {
			client.DestroyVolumeSnapshot(updated.Id)
			client.EradicateVolumeSnapshot(updated.Id)
		})
	})
}
