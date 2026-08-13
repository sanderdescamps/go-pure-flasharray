package flashclient_test

import (
	"testing"

	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func TestPods(t *testing.T) {
	client, teardown := setupTestClient(t)
	t.Cleanup(teardown)

	t.Run("pods_create", func(t *testing.T) {
		name := NewRandomName("pod", 8)

		pod, err := client.CreatePod(name, flashclient.PodPostBody{
			QuotaLimit: toPtr[int64](2000),
		})
		if err != nil {
			t.Fatalf("Failed to create pod: %v", err)
		}
		t.Logf("Pod created successfully: name=%s, Id=%s", pod.Name, pod.Id)

		podByName, err := client.GetPodByName(name)
		if err != nil {
			t.Fatalf("Failed to get pod by name: %v", err)
		} else if podByName.Name != name {
			t.Fatalf("Expected pod name %s, got %s", name, podByName.Name)
		}
		t.Logf("Pod retrieved successfully by name: name=%s, Id=%s", podByName.Name, podByName.Id)

		pod, err = client.GetPod(pod.Id)
		if err != nil {
			t.Fatalf("Failed to get pod by ID: %v", err)
		} else if pod.Name != name {
			t.Fatalf("Expected pod name %s, got %s", name, pod.Name)
		}
		t.Logf("Pod retrieved successfully by ID: name=%s, Id=%s", pod.Name, pod.Id)

		allPods, err := client.GetPods()
		if err != nil {
			t.Fatalf("Failed to get pods: %v", err)
		}
		if lenPods := len(allPods); lenPods > 0 {
			t.Logf("%d Pods found", lenPods)
		} else {
			t.Errorf("Expected at least 1 pod")
		}

		err = client.DestroyPod(pod.Id)
		if err != nil {
			t.Fatalf("Failed to destroy pod by ID: %v", err)
		}
		t.Logf("Pod destroyed successfully by ID: name=%s, Id=%s", pod.Name, pod.Id)

		err = client.EradicatePod(pod.Id)
		if err != nil {
			t.Fatalf("Failed to eradicate pod by ID: %v", err)
		}
		t.Logf("Pod eradicated successfully by ID: name=%s, Id=%s", pod.Name, pod.Id)
	})

	t.Run("pod_get_not_found", func(t *testing.T) {
		_, err := client.GetPod("00000000-0000-0000-0000-000000000000")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent pod, got nil")
		}
		t.Logf("Got expected error for non-existent pod ID: %v", err)
	})

	t.Run("pod_get_by_name_not_found", func(t *testing.T) {
		_, err := client.GetPodByName("non-existent-pod-name")
		if err == nil {
			t.Fatalf("Expected error when getting non-existent pod by name, got nil")
		}
		t.Logf("Got expected error for non-existent pod name: %v", err)
	})

	t.Run("pod_update_rename", func(t *testing.T) {
		name := NewRandomName("pod", 8)
		pod, err := client.CreatePod(name, flashclient.PodPostBody{})
		if err != nil {
			t.Fatalf("Failed to create pod: %v", err)
		}
		t.Cleanup(func() {
			client.DestroyPod(pod.Id)
			client.EradicatePod(pod.Id)
		})

		newName := NewRandomName("pod", 8)
		updated, err := client.UpdatePod(pod.Id, flashclient.PodPatchBody{Name: toPtr(newName)})
		if err != nil {
			t.Fatalf("Failed to rename pod: %v", err)
		}
		if updated.Name != newName {
			t.Errorf("Expected pod name %s, got %s", newName, updated.Name)
		}
		t.Logf("Pod renamed: %s -> %s", name, updated.Name)

		retrieved, err := client.GetPodByName(newName)
		if err != nil {
			t.Fatalf("Failed to get pod by new name: %v", err)
		}
		if retrieved.Name != newName {
			t.Errorf("Expected pod name %s, got %s", newName, retrieved.Name)
		}
	})

}
