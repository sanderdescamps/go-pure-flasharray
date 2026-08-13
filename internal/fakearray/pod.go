package fakearray

import (
	"fmt"

	"github.com/google/uuid"
	faclient "github.com/sanderdescamps/go-purefa"
)

func NewPodPost(name string, podPost faclient.PodPostBody) faclient.Pod {

	newPod := faclient.Pod{
		FixedReference: faclient.FixedReference{
			Id:   uuid.New().String(),
			Name: name,
		},
	}

	if len(podPost.FailoverPreferences) != 0 {
		newPod.FailoverPreferences = podPost.FailoverPreferences
	}
	if podPost.QuotaLimit != nil {
		newPod.QuotaLimit = *podPost.QuotaLimit
	}

	if podPost.Source != nil {
		newPod.Source = *podPost.Source
	}

	return newPod
}

func (array *Array) GetPod(id string) (*faclient.Pod, error) {
	for _, pod := range array.Pods {
		if pod.Id == id {
			return pod, nil
		}
	}
	return nil, fmt.Errorf("pod with ID %s not found", id)
}

func (array *Array) GetPods() []faclient.Pod {
	pods := make([]faclient.Pod, len(array.Pods))
	for i, pod := range array.Pods {
		pods[i] = *pod
	}
	return pods
}

func (array *Array) GetPodsByName(name string) (*faclient.Pod, error) {
	for _, pod := range array.Pods {
		if pod.Name == name {
			return pod, nil
		}
	}
	return nil, fmt.Errorf("pod with name %s not found", name)
}

func (array *Array) AddPod(pod faclient.Pod) (*faclient.Pod, error) {
	for _, p := range array.Pods {
		if p.Name == pod.Name {
			return nil, fmt.Errorf("pod with name %s already exists: %w", pod.Name, ErrAlreadyExists)
		}
	}
	if pod.Id == "" {
		pod.Id = uuid.New().String()
	}
	array.Pods = append(array.Pods, &pod)
	return &pod, nil
}

func (array *Array) CreatePod(name string, podPost faclient.PodPostBody) (*faclient.Pod, error) {
	newPod := NewPodPost(name, podPost)
	return array.AddPod(newPod)
}

func (array *Array) UpdatePod(id string, patch faclient.PodPatchBody) (*faclient.Pod, error) {
	for i, pod := range array.Pods {
		if pod.Id == id {
			if patch.Name != nil {
				pod.Name = *patch.Name
			}
			if patch.Destroyed != nil {
				pod.Destroyed = *patch.Destroyed
			}
			if patch.Mediator != nil {
				pod.Mediator = *patch.Mediator
			}
			if patch.RequestedPromotionState != nil {
				pod.RequestedPromotionState = *patch.RequestedPromotionState
			}
			array.Pods[i] = pod
			return pod, nil
		}
	}
	return nil, fmt.Errorf("pod with ID %s not found", id)
}

// Destroy pod
func (array *Array) DestroyPod(id string) error {
	_, err := array.UpdatePod(id, faclient.PodPatchBody{Destroyed: new(bool)})
	return err
}

func (array *Array) EradicatePod(id string) error {
	for i, pod := range array.Pods {
		if pod.Id == id {
			array.Pods = append(array.Pods[:i], array.Pods[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("pod with ID %s not found", id)
}
