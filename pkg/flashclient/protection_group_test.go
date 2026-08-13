package flashclient_test

import (
	"fmt"
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func TestProtectionGroups(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("protection_groups_list", func(t *testing.T) {
		pgs, err := client.GetProtectionGroups()
		if err != nil {
			t.Fatalf("Failed to get protection groups: %v", err)
		}
		if len(pgs) < 1 {
			t.Errorf("Expected at least 1 protection group, but got %d", len(pgs))
		}
		t.Logf("%d protection groups found", len(pgs))
	})

	t.Run("protection_group_lifecycle", func(t *testing.T) {
		name := NewRandomName("pg", 8)
		pg, err := client.CreateProtectionGroup(name)
		if err != nil {
			t.Fatalf("Failed to create protection group: %v", err)
		}
		if pg.Name != name {
			t.Errorf("Expected name %s, got %s", name, pg.Name)
		}
		t.Logf("Protection group created: name=%s, id=%s", pg.Name, pg.Id)

		retrieved, err := client.GetProtectionGroup(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group by ID: %v", err)
		}
		if retrieved.Id != pg.Id {
			t.Errorf("Expected ID %s, got %s", pg.Id, retrieved.Id)
		}
		t.Logf("Protection group retrieved by ID: name=%s", retrieved.Name)

		retrievedByName, err := client.GetProtectionGroupByName(pg.Name)
		if err != nil {
			t.Fatalf("Failed to get protection group by name: %v", err)
		}
		if retrievedByName.Id != pg.Id {
			t.Errorf("Expected ID %s from name lookup, got %s", pg.Id, retrievedByName.Id)
		}
		t.Logf("Protection group retrieved by name: name=%s", retrievedByName.Name)

		newName := NewRandomName("pg", 8)
		updated, err := client.UpdateProtectionGroup(pg.Id, flashclient.ProtectionGroupPatchBody{Name: &newName})
		if err != nil {
			t.Fatalf("Failed to rename protection group: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Expected renamed name %s, got %s", newName, updated.Name)
		}
		t.Logf("Protection group renamed: %s -> %s", name, updated.Name)

		err = client.DestroyProtectionGroup(pg.Id)
		if err != nil {
			t.Fatalf("Failed to destroy protection group: %v", err)
		}
		t.Logf("Protection group destroyed: id=%s", pg.Id)

		err = client.EradicateProtectionGroup(pg.Id)
		if err != nil {
			t.Fatalf("Failed to eradicate protection group: %v", err)
		}
		t.Logf("Protection group eradicated: id=%s", pg.Id)
	})

	t.Run("protection_group_host_members", func(t *testing.T) {
		pgName := NewRandomName("pg", 8)
		pg, err := client.CreateProtectionGroup(pgName)
		if err != nil {
			t.Fatalf("Failed to create protection group: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyProtectionGroup(pg.Id)
			client.EradicateProtectionGroup(pg.Id)
		})

		hostName := NewRandomName("host", 8)
		host, err := client.CreateHost(hostName, flashclient.HostPostBody{
			IQNs: []string{fmt.Sprintf("iqn.2026-example.com:%s", NewString(8))},
		})
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}
		t.Cleanup(func() { client.DeleteHost(host.Name) })

		err = client.AddHostToProtectionGroup(pg.Id, host.Name)
		if err != nil {
			t.Fatalf("Failed to add host to protection group: %v", err)
		}
		t.Logf("Host %s added to protection group %s", host.Name, pg.Name)

		members, err := client.GetProtectionGroupHostMembers(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group host members: %v", err)
		}
		found := false
		for _, m := range members {
			if m.Member.Name == host.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected host %s in protection group host members", host.Name)
		}
		t.Logf("Host %s confirmed in protection group members", host.Name)

		allMembers, err := client.GetAllProtectionGroupHostMembers()
		if err != nil {
			t.Fatalf("Failed to get all protection group host members: %v", err)
		}
		t.Logf("%d total protection group host members", len(allMembers))

		err = client.RemoveHostFromProtectionGroup(pg.Id, host.Name)
		if err != nil {
			t.Fatalf("Failed to remove host from protection group: %v", err)
		}
		t.Logf("Host %s removed from protection group %s", host.Name, pg.Name)

		members, err = client.GetProtectionGroupHostMembers(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group host members after removal: %v", err)
		}
		for _, m := range members {
			if m.Member.Name == host.Name {
				t.Errorf("Host %s should no longer be a member of protection group %s", host.Name, pg.Name)
			}
		}
	})

	t.Run("protection_group_host_group_members", func(t *testing.T) {
		pgName := NewRandomName("pg", 8)
		pg, err := client.CreateProtectionGroup(pgName)
		if err != nil {
			t.Fatalf("Failed to create protection group: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyProtectionGroup(pg.Id)
			client.EradicateProtectionGroup(pg.Id)
		})

		hgName := NewRandomName("hg", 8)
		hg, err := client.CreateHostGroup(hgName)
		if err != nil {
			t.Fatalf("Failed to create host group: %v", err)
		}
		t.Cleanup(func() { client.DeleteHostGroup(hg.Name) })

		err = client.AddHostGroupToProtectionGroup(pg.Id, hg.Name)
		if err != nil {
			t.Fatalf("Failed to add host group to protection group: %v", err)
		}
		t.Logf("Host group %s added to protection group %s", hg.Name, pg.Name)

		members, err := client.GetProtectionGroupHostGroupMembers(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group host group members: %v", err)
		}
		found := false
		for _, m := range members {
			if m.Member.Name == hg.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected host group %s in protection group host group members", hg.Name)
		}
		t.Logf("Host group %s confirmed in protection group members", hg.Name)

		allMembers, err := client.GetAllProtectionGroupHostGroupMembers()
		if err != nil {
			t.Fatalf("Failed to get all protection group host group members: %v", err)
		}
		t.Logf("%d total protection group host group members", len(allMembers))

		err = client.RemoveHostGroupFromProtectionGroup(pg.Id, hg.Name)
		if err != nil {
			t.Fatalf("Failed to remove host group from protection group: %v", err)
		}
		t.Logf("Host group %s removed from protection group %s", hg.Name, pg.Name)

		members, err = client.GetProtectionGroupHostGroupMembers(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group host group members after removal: %v", err)
		}
		for _, m := range members {
			if m.Member.Name == hg.Name {
				t.Errorf("Host group %s should no longer be a member of protection group %s", hg.Name, pg.Name)
			}
		}
	})

	t.Run("protection_group_volume_members", func(t *testing.T) {
		pgName := NewRandomName("pg", 8)
		pg, err := client.CreateProtectionGroup(pgName)
		if err != nil {
			t.Fatalf("Failed to create protection group: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyProtectionGroup(pg.Id)
			client.EradicateProtectionGroup(pg.Id)
		})

		volName := NewRandomName("vol", 8)
		vol, err := client.CreateVolume(volName, flashclient.VolumePost{Provisioned: 1048576})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
		t.Cleanup(func() {
			client.DeleteVolume(vol.Id)
			client.EradicateVolume(vol.Id)
		})

		err = client.AddVolumeToProtectionGroup(pg.Id, vol.Id)
		if err != nil {
			t.Fatalf("Failed to add volume to protection group: %v", err)
		}
		t.Logf("Volume %s added to protection group %s", vol.Name, pg.Name)

		members, err := client.GetProtectionGroupVolumeMembers(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group volume members: %v", err)
		}
		found := false
		for _, m := range members {
			if m.Member.Id == vol.Id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected volume %s in protection group volume members", vol.Name)
		}
		t.Logf("Volume %s confirmed in protection group members", vol.Name)

		allMembers, err := client.GetAllProtectionGroupVolumeMembers()
		if err != nil {
			t.Fatalf("Failed to get all protection group volume members: %v", err)
		}
		t.Logf("%d total protection group volume members", len(allMembers))

		err = client.RemoveVolumeFromProtectionGroup(pg.Id, vol.Id)
		if err != nil {
			t.Fatalf("Failed to remove volume from protection group: %v", err)
		}
		t.Logf("Volume %s removed from protection group %s", vol.Name, pg.Name)

		members, err = client.GetProtectionGroupVolumeMembers(pg.Id)
		if err != nil {
			t.Fatalf("Failed to get protection group volume members after removal: %v", err)
		}
		for _, m := range members {
			if m.Member.Id == vol.Id {
				t.Errorf("Volume %s should no longer be a member of protection group %s", vol.Name, pg.Name)
			}
		}
	})
}
