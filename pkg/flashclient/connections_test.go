package flashclient_test

import (
	"fmt"
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func TestConnections(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("connections_list", func(t *testing.T) {
		connections, err := client.GetConnections()
		if err != nil {
			t.Fatalf("Failed to get connections: %v", err)
		}
		if len(connections) == 0 {
			t.Errorf("Expected at least 1 connection")
		}
		t.Logf("%d connections found", len(connections))

		for _, c := range connections {
			if c.Host.Name == "" && c.HostGroup.Name == "" {
				t.Errorf("Connection has empty host name and host group: %+v", c)
			}
			if c.Volume.Name == "" {
				t.Errorf("Connection has empty volume name: %+v", c)
			}
		}
	})

	t.Run("host_connection_lifecycle", func(t *testing.T) {
		hostName := NewRandomName("host", 8)
		host, err := client.CreateHost(hostName, flashclient.HostPostBody{
			IQNs: []string{fmt.Sprintf("iqn.2026-example.com:%s", NewString(8))},
		})
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}
		t.Cleanup(func() { client.DeleteHost(host.Name) })

		volumeName := NewRandomName("vol", 8)
		volume, err := client.CreateVolume(volumeName, flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		connections, err := client.CreateHostConnections([]string{hostName}, []string{volume.Id}, flashclient.ConnectionPostBody{})
		if err != nil {
			t.Fatalf("Failed to create host connection: %v", err)
		}
		if len(connections) != 1 {
			t.Fatalf("Expected 1 connection, got %d", len(connections))
		}
		if connections[0].Host.Name != hostName {
			t.Errorf("Expected host name %s, got %s", hostName, connections[0].Host.Name)
		}
		if connections[0].Volume.Id != volume.Id {
			t.Errorf("Expected volume ID %s, got %s", volume.Id, connections[0].Volume.Id)
		}
		t.Logf("Host connection created: host=%s, volume=%s, lun=%v", hostName, volumeName, connections[0].Lun)

		err = client.DeleteHostConnections([]string{hostName}, []string{volume.Id})
		if err != nil {
			t.Fatalf("Failed to delete host connection: %v", err)
		}
		t.Logf("Host connection deleted: host=%s, volume=%s", hostName, volumeName)

		allConns, err := client.GetConnections()
		if err != nil {
			t.Fatalf("Failed to get connections after delete: %v", err)
		}
		for _, c := range allConns {
			if c.Host.Name == hostName && c.Volume.Id == volume.Id {
				t.Errorf("Expected connection host=%s volume=%s to be deleted", hostName, volume.Id)
			}
		}
	})

	t.Run("host_group_connection_lifecycle", func(t *testing.T) {
		hostName := NewRandomName("host", 8)
		host, err := client.CreateHost(hostName, flashclient.HostPostBody{
			IQNs: []string{fmt.Sprintf("iqn.2026-example.com:%s", NewString(8))},
		})
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}
		t.Cleanup(func() { client.DeleteHost(host.Name) })

		groupName := NewRandomName("hg", 8)
		group, err := client.CreateHostGroup(groupName)
		if err != nil {
			t.Fatalf("Failed to create host group: %v", err)
		}
		t.Cleanup(func() { client.DeleteHostGroup(group.Name) })

		_, err = client.AddHostGroupMembers(groupName, []string{hostName})
		if err != nil {
			t.Fatalf("Failed to add host to host group: %v", err)
		}

		volumeName := NewRandomName("vol", 8)
		volume, err := client.CreateVolume(volumeName, flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyVolume(volume.Id)
			client.EradicateVolume(volume.Id)
		})

		connections, err := client.CreateHostGroupConnections([]string{groupName}, []string{volume.Id}, flashclient.ConnectionPostBody{})
		if err != nil {
			t.Fatalf("Failed to create host group connection: %v", err)
		}
		if len(connections) == 0 {
			t.Fatalf("Expected at least 1 connection, got 0")
		}
		for _, c := range connections {
			if c.HostGroup.Name != groupName {
				t.Errorf("Expected host group name %s, got %s", groupName, c.HostGroup.Name)
			}
			if c.Volume.Id != volume.Id {
				t.Errorf("Expected volume ID %s, got %s", volume.Id, c.Volume.Id)
			}
		}
		t.Logf("Host group connection created: group=%s, volume=%s, connections=%d", groupName, volumeName, len(connections))

		err = client.DeleteHostGroupConnections([]string{groupName}, []string{volume.Id})
		if err != nil {
			t.Fatalf("Failed to delete host group connection: %v", err)
		}
		t.Logf("Host group connection deleted: group=%s, volume=%s", groupName, volumeName)

		allConns, err := client.GetConnections()
		if err != nil {
			t.Fatalf("Failed to get connections after delete: %v", err)
		}
		for _, c := range allConns {
			if c.HostGroup.Name == groupName && c.Volume.Id == volume.Id {
				t.Errorf("Expected connection group=%s volume=%s to be deleted", groupName, volume.Id)
			}
		}
	})
}
