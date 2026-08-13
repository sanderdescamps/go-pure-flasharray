package flashclient_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func TestHosts(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("host_list", func(t *testing.T) {
		hosts, err := client.GetHosts()
		if err != nil {
			t.Fatalf("Failed to get hosts: %v", err)
		}
		if lenHosts := len(hosts); lenHosts > 0 {
			t.Logf("%d Hosts found", lenHosts)
		} else {
			t.Errorf("Expected at least 1 host")
		}
	})

	t.Run("host_create", func(t *testing.T) {
		hostPost := flashclient.HostPostBody{
			IQNs: []string{fmt.Sprintf("iqn.2026-example.com:%s", NewString(8))},
		}
		name := NewRandomName("host", 8)
		host, err := client.CreateHost(name, hostPost)
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		} else if host.Name != name {
			t.Fatalf("Expected host name %s, got %s", name, host.Name)
		}
		t.Logf("Host created successfully: name=%s, IQNs=%s", host.Name, strings.Join(host.IQNs, ","))

		host, err = client.GetHost(name)
		if err != nil {
			t.Fatalf("Failed to get host by name: %v", err)
		} else if host.Name != name {
			t.Fatalf("Expected host name %s, got %s", name, host.Name)
		}
		t.Logf("Host retrieved successfully by name: name=%s, IQNs=%s", host.Name, strings.Join(host.IQNs, ","))

		err = client.DeleteHost(name)
		if err != nil {
			t.Fatalf("Failed to delete host by name: %v", err)
		}
		t.Logf("Host deleted successfully by name: name=%s", name)
	})

	t.Run("host_lifecycle", func(t *testing.T) {
		name := NewRandomName("host", 8)
		host, err := client.CreateHost(name, flashclient.HostPostBody{
			IQNs: []string{
				fmt.Sprintf("iqn.2026-08.com-purestorage:%s", NewString(8)),
			},
		})
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}
		if host.Name != name {
			t.Fatalf("Expected name %s, got %s", name, host.Name)
		}
		t.Logf("Host created: name=%s", host.Name)

		retrieved, err := client.GetHost(name)
		if err != nil {
			t.Fatalf("Failed to get host by name: %v", err)
		}
		if retrieved.Name != name {
			t.Fatalf("Expected name %s, got %s", name, retrieved.Name)
		}
		t.Logf("Host retrieved by name: name=%s", retrieved.Name)

		newName := NewRandomName("host", 8)
		updated, err := client.UpdateHost(name, flashclient.HostPatchBody{Name: &newName})
		if err != nil {
			t.Fatalf("Failed to rename host: %v", err)
		}
		if updated.Name != newName {
			t.Fatalf("Expected renamed name %s, got %s", newName, updated.Name)
		}
		t.Logf("Host renamed: %s -> %s", name, updated.Name)

		err = client.DeleteHost(newName)
		if err != nil {
			t.Fatalf("Failed to delete host: %v", err)
		}
		t.Logf("Host deleted: name=%s", newName)
	})

	t.Run("host_get_not_found", func(t *testing.T) {
		_, err := client.GetHost("non-existent-host-name")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent host, got nil")
		}
		t.Logf("Got expected error for non-existent host name: %v", err)
	})
}
