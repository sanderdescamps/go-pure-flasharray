package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func TestProtectionGroupSnapshots(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("protection_group_snapshot_lifecycle", func(t *testing.T) {
		pgName := NewRandomName("pg", 8)

		// Create a protection group to snapshot
		t.Logf("Creating protection group %q", pgName)
		pg, err := client.CreateProtectionGroup(pgName)
		if err != nil {
			t.Fatalf("failed to create protection group %q: %v", pgName, err)
		}
		t.Cleanup(func() {
			if err := client.DestroyProtectionGroup(pg.Id); err != nil {
				t.Logf("warning: failed to destroy protection group %q: %v", pgName, err)
			}
			if err := client.EradicateProtectionGroup(pg.Id); err != nil {
				t.Logf("warning: failed to eradicate protection group %q: %v", pgName, err)
			}
		})
		t.Logf("Protection group created: Name=%s, ID=%s", pg.Name, pg.Id)

		// Create a snapshot of the protection group
		snap, err := client.CreateProtectionGroupSnapshot(pg.Id, flashclient.ProtectionGroupSnapshotPostBody{})
		if err != nil {
			t.Fatalf("failed to create protection group snapshot: %v", err)
		}
		t.Logf("Protection group snapshot created: Name=%s, ID=%s", snap.Name, snap.Id)

		// List protection group snapshots and verify the new one is present
		t.Logf("Listing protection group snapshots for %q", pgName)
		snaps, err := client.GetProtectionGroupSnapshots()
		if err != nil {
			t.Fatalf("failed to list protection group snapshots: %v", err)
		}
		found := false
		for _, s := range snaps {
			if s.Name == snap.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("newly created snapshot %q not found in list", snap.Name)
		}

		// List protection group snapshots for the specific source protection group and verify the new one is present
		t.Logf("Listing protection group snapshots for %q", pgName)
		snaps, err = client.GetProtectionGroupSnapshotsForSource(pg.Id)
		if err != nil {
			t.Fatalf("failed to list protection group snapshots: %v", err)
		}
		found = false
		for _, s := range snaps {
			if s.Name == snap.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("newly created snapshot %q not found in list", snap.Name)
		}

		// Delete the snapshot explicitly (before cleanup runs) to verify deletion works
		t.Logf("Deleting protection group snapshot %q", snap.Name)
		if err := client.DestroyProtectionGroupSnapshot(snap.Id); err != nil {
			t.Fatalf("failed to delete protection group snapshot Name=%s, ID=%s: %v", snap.Name, snap.Id, err)
		}
		if err := client.EradicateProtectionGroupSnapshot(snap.Id); err != nil {
			t.Fatalf("failed to delete protection group snapshot Name=%s, ID=%s: %v", snap.Name, snap.Id, err)
		}

		// Verify the snapshot no longer exists
		t.Logf("Verifying protection group snapshot Name=%s, ID=%s is deleted", snap.Name, snap.Id)
		_, err = client.GetProtectionGroupSnapshot(snap.Id)
		if err == nil {
			t.Errorf("expected error when getting deleted snapshot Name=%s, ID=%s, but got none", snap.Name, snap.Id)
		}
	})
}
