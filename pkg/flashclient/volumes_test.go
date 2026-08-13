package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func TestVolumes(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("volumes_list", func(t *testing.T) {
		volumes, err := client.GetVolumes()
		if err != nil {
			t.Fatalf("Failed to get volumes: %v", err)
		}
		if lenVolumes := len(volumes); lenVolumes > 0 {
			t.Logf("%d Volumes found", lenVolumes)
		} else {
			t.Errorf("Expected at least 1 volume, got %d", len(volumes))
		}
	})

	t.Run("volume_create", func(t *testing.T) {
		volumePost := flashclient.VolumePost{
			Provisioned: 1024,
			QoS: &flashclient.Qos{
				BandwidthLimit: toPtr[int64](2000),
			},
		}

		name := NewRandomName("volume", 8)
		volume, err := client.CreateVolume(name, volumePost)
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Logf("Volume created successfully: name=%s, Id=%s", volume.Name, volume.Id)

		volume, err = client.GetVolumeByName(name)
		if err != nil {
			t.Fatalf("Failed to get volume by name: %v", err)
		} else if volume.Name != name {
			t.Fatalf("Expected volume name %s, got %s", name, volume.Name)
		}
		t.Logf("Volume retrieved successfully by name: name=%s, Id=%s", volume.Name, volume.Id)

		volume, err = client.GetVolume(volume.Id)
		if err != nil {
			t.Fatalf("Failed to get volume by ID: %v", err)
		} else if volume.Name != name {
			t.Fatalf("Expected volume name %s, got %s", name, volume.Name)
		}
		t.Logf("Volume retrieved successfully by ID: name=%s, Id=%s", volume.Name, volume.Id)

		err = client.DeleteVolume(volume.Id)
		if err != nil {
			t.Fatalf("Failed to delete volume by ID: %v", err)
		}
		t.Logf("Volume deleted successfully by ID: name=%s, Id=%s", volume.Name, volume.Id)

		err = client.EradicateVolume(volume.Id)
		if err != nil {
			t.Fatalf("Failed to eradicate volume by ID: %v", err)
		}
		t.Logf("Volume eradicated successfully by ID: name=%s, Id=%s", volume.Name, volume.Id)
	})

	t.Run("volume_get_not_found", func(t *testing.T) {
		_, err := client.GetVolume("00000000-0000-0000-0000-000000000000")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent volume, got nil")
		}
		t.Logf("Got expected error for non-existent volume ID: %v", err)
	})

	t.Run("volume_get_by_name_not_found", func(t *testing.T) {
		_, err := client.GetVolumeByName("non-existent-volume-name")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent volume by name, got nil")
		}
		t.Logf("Got expected error for non-existent volume name: %v", err)
	})

	t.Run("volume_update_resize", func(t *testing.T) {
		const initialSize int64 = 1048576
		const resizedSize int64 = 2097152

		name := NewRandomName("volume", 8)
		volume, err := client.CreateVolume(name, flashclient.VolumePost{Provisioned: initialSize})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DeleteVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		_, err = client.UpdateVolume(volume.Id, flashclient.VolumePatch{
			Provisioned: toPtr(resizedSize),
		})
		if err != nil {
			t.Fatalf("Failed to resize volume: %v", err)
		}

		updated, err := client.GetVolume(volume.Id)
		if err != nil {
			t.Fatalf("Failed to get volume after resize: %v", err)
		}
		if updated.Provisioned != resizedSize {
			t.Errorf("Expected provisioned size %d, got %d", resizedSize, updated.Provisioned)
		}
		t.Logf("Volume resized successfully: name=%s, provisioned=%d", updated.Name, updated.Provisioned)
	})

	t.Run("volume_update_qos", func(t *testing.T) {
		name := NewRandomName("volume", 8)
		volume, err := client.CreateVolume(name, flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DeleteVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		const newBandwidthLimit int64 = 4096
		_, err = client.UpdateVolume(volume.Id, flashclient.VolumePatch{
			QoS: &flashclient.Qos{BandwidthLimit: toPtr(newBandwidthLimit)},
		})
		if err != nil {
			t.Fatalf("Failed to update volume QoS: %v", err)
		}

		updated, err := client.GetVolume(volume.Id)
		if err != nil {
			t.Fatalf("Failed to get volume after QoS update: %v", err)
		}
		if updated.QoS.BandwidthLimit == nil || *updated.QoS.BandwidthLimit != newBandwidthLimit {
			t.Errorf("Expected bandwidth limit %d, got %v", newBandwidthLimit, updated.QoS.BandwidthLimit)
		}
		t.Logf("Volume QoS updated successfully: name=%s, bandwidth_limit=%d", updated.Name, *updated.QoS.BandwidthLimit)
	})

	t.Run("volume_update_rename", func(t *testing.T) {
		name := NewRandomName("volume", 8)
		volume, err := client.CreateVolume(name, flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DeleteVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		newName := NewRandomName("volume", 8)
		updated, err := client.UpdateVolume(volume.Id, flashclient.VolumePatch{Name: toPtr(newName)})
		if err != nil {
			t.Fatalf("Failed to rename volume: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Expected volume name %s, got %s", newName, updated.Name)
		}
		t.Logf("Volume renamed: %s -> %s", name, updated.Name)

		_, err = client.GetVolumeByName(name)
		if err == nil {
			t.Errorf("Expected error when getting volume by old name %s after rename, got nil", name)
		}
	})
}
