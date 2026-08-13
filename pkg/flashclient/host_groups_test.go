package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func TestHostGroups(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("host_groups_list", func(t *testing.T) {
		hostGroups, err := client.GetHostGroups()
		if err != nil {
			t.Fatalf("Failed to get host groups: %v", err)
		}
		if lenHostGroups := len(hostGroups); lenHostGroups > 0 {
			t.Logf("%d Host groups found", lenHostGroups)
		} else {
			t.Errorf("Expected at least 1 host group")
		}
	})

	t.Run("host_group_update_rename", func(t *testing.T) {
		name := NewRandomName("hostgroup", 8)
		hostGroup, err := client.CreateHostGroup(name)
		if err != nil {
			t.Fatalf("Failed to create host group: %v", err)
		}

		newName := NewRandomName("hostgroup", 8)
		hostGroup, err = client.UpdateHostGroup(name, flashclient.HostGroupPatchBody{Name: newName})
		if err != nil {
			t.Fatalf("Failed to rename host group: %v", err)
		}

		if hostGroup.Name != newName {
			t.Errorf("Expected host group name %s after rename, got %s", newName, hostGroup.Name)
		}
		t.Logf("Host group renamed successfully: %s -> %s", name, hostGroup.Name)

		retrieved, err := client.GetHostGroup(newName)
		if err != nil {
			t.Fatalf("Failed to get host group by new name: %v", err)
		}
		if retrieved.Name != newName {
			t.Errorf("Expected host group name %s, got %s", newName, retrieved.Name)
		}

		_, err = client.GetHostGroup(name)
		if err == nil {
			t.Errorf("Expected error when getting host group by old name %s after rename, got nil", name)
		}

		err = client.DeleteHostGroup(newName)
		if err != nil {
			t.Fatalf("Failed to delete host group by new name: %v", err)
		}
		t.Logf("Host group deleted successfully: name=%s", newName)
	})

	t.Run("host_groups_create", func(t *testing.T) {
		name := NewRandomName("hostgroup", 8)
		hostGroup, err := client.CreateHostGroup(name)
		if err != nil {
			t.Fatalf("Failed to create host group: %v", err)
		} else if hostGroup.Name != name {
			t.Fatalf("Expected host group name %s, got %s", name, hostGroup.Name)
		}
		t.Cleanup(func() { client.DeleteHostGroup(name) })
		t.Logf("Host group created successfully: name=%s", hostGroup.Name)

		hostGroup, err = client.GetHostGroup(name)
		if err != nil {
			t.Fatalf("Failed to get host group by name: %v", err)
		} else if hostGroup.Name != name {
			t.Fatalf("Expected host group name %s, got %s", name, hostGroup.Name)
		}
		t.Logf("Host group retrieved successfully by name: name=%s", hostGroup.Name)

		err = client.DeleteHostGroup(name)
		if err != nil {
			t.Fatalf("Failed to delete host group by name: %v", err)
		}
		t.Logf("Host group deleted successfully by name: name=%s", name)
	})

	t.Run("host_group_members", func(t *testing.T) {
		hostName := NewRandomName("host", 8)
		host, err := client.CreateHost(hostName, flashclient.HostPostBody{})
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}
		t.Logf("Host created successfully: name=%s", host.Name)

		groupName := NewRandomName("hostgroup", 8)
		hostGroup, err := client.CreateHostGroup(groupName)
		if err != nil {
			t.Fatalf("Failed to create host group: %v", err)
		}
		t.Logf("Host group created successfully: name=%s", hostGroup.Name)

		//add member to host group
		members, err := client.AddHostGroupMembers(hostGroup.Name, []string{host.Name})
		if err != nil {
			t.Fatalf("Failed to add host to host group: %v", err)
		}
		if len(members) == 0 {
			t.Fatalf("Expected at least 1 member after adding host to host group")
		}
		t.Logf("Host %s added to host group %s", host.Name, hostGroup.Name)

		// check membership
		members, err = client.GetHostGroupMembers([]string{hostGroup.Name}, nil)
		if err != nil {
			t.Fatalf("Failed to get host group members: %v", err)
		}
		found := false
		for _, m := range members {
			if m.Member.Name == host.Name && m.Group.Name == hostGroup.Name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Expected host %s to be a member of host group %s", host.Name, hostGroup.Name)
		}
		t.Logf("Host %s confirmed as member of host group %s", host.Name, hostGroup.Name)

		// remove member from host group
		err = client.DeleteHostGroupMembers(groupName, []string{hostName})
		if err != nil {
			t.Fatalf("Failed to remove host from host group: %v", err)
		}
		t.Logf("Host %s removed from host group %s", hostName, groupName)

		// check membership after removal
		members, err = client.GetHostGroupMembers([]string{groupName}, []string{hostName})
		if err != nil {
			t.Fatalf("Failed to get host group members after removal: %v", err)
		}
		for _, m := range members {
			if m.Member.Name == hostName && m.Group.Name == groupName {
				t.Fatalf("Expected host %s to no longer be a member of host group %s", hostName, groupName)
			}
		}
		t.Logf("Host %s confirmed no longer a member of host group %s", hostName, groupName)

		//cleanup
		err = client.DeleteHost(hostName)
		if err != nil {
			t.Fatalf("Failed to delete host: %v", err)
		}
		t.Logf("Host %s deleted successfully", hostName)

		err = client.DeleteHostGroup(groupName)
		if err != nil {
			t.Fatalf("Failed to delete host group: %v", err)
		}
		t.Logf("Host group %s deleted successfully", groupName)
	})

	t.Run("host_group_get_not_found", func(t *testing.T) {
		_, err := client.GetHostGroup("non-existent-host-group-name")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent host group, got nil")
		}
		t.Logf("Got expected error for non-existent host group: %v", err)
	})
}
