package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func InitPodRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/pods").Handler(GetPodsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/pods").Handler(PostPodHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/pods").Handler(PatchPodHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/pods").Handler(DeletePodHandler(array, logger))
}

func GetPodsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pods := []flashclient.Pod{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				pod, err := array.GetPod(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get pod", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Pod with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found pod by ID", "id", id, "name", pod.Name)
				pods = append(pods, *pod)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				pod, err := array.GetPodsByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get pod", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Pod with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found pod by name", "name", name, "id", pod.Id)
				pods = append(pods, *pod)
			}
		} else {
			logger.DebugContext(r.Context(), "Getting all pods")
			pods = array.GetPods()
		}

		logger.DebugContext(r.Context(), "Returning pods", "count", len(pods))
		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(pods)
		json.NewEncoder(w).Encode(data)
	}
}

func PostPodHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		podNames := splitQueryParam(r.URL.Query().Get("names"))
		if len(podNames) < 1 || podNames[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		podPost := flashclient.PodPostBody{}
		err := json.NewDecoder(r.Body).Decode(&podPost)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		podsCreated := []flashclient.Pod{}
		for _, name := range podNames {
			pod, err := array.CreatePod(name, podPost)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to create pod", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to add pod: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Created pod", "name", pod.Name, "id", pod.Id)
			podsCreated = append(podsCreated, *pod)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(podsCreated))
	}
}

func PatchPodHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		podIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				pod, err := array.GetPod(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get pod", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Pod with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found pod by ID for patch", "id", id, "name", pod.Name)
				podIds = append(podIds, pod.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				pod, err := array.GetPodsByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get pod", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Pod with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found pod by name for patch", "name", name, "id", pod.Id)
				podIds = append(podIds, pod.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		var podPatch flashclient.PodPatchBody
		err := json.NewDecoder(r.Body).Decode(&podPatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(podIds) > 1 && podPatch.Name != nil {
			httpJsonError(w, "Cannot specify multiple names in the query parameter when 'name' field is set in the request body", http.StatusBadRequest)
			return
		}

		pods := []flashclient.Pod{}
		for _, id := range podIds {
			pod, err := array.UpdatePod(id, podPatch)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to update pod", "id", id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to update pod: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated pod", "id", pod.Id, "name", pod.Name)
			pods = append(pods, *pod)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(pods))
	}
}

func DeletePodHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		podIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				pod, err := array.GetPod(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get pod", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Pod with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found pod by ID for deletion", "id", id, "name", pod.Name)
				podIds = append(podIds, pod.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				pod, err := array.GetPodsByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get pod", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Pod with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found pod by name for deletion", "name", name, "id", pod.Id)
				podIds = append(podIds, pod.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, id := range podIds {
			err := array.EradicatePod(id)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to delete pod", "id", id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to delete pod: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Deleted pod", "id", id)
		}

		w.WriteHeader(http.StatusOK)
	}
}
