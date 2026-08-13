package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func InitVolumeRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/volumes").Handler(GetVolumesHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/volumes").Handler(PatchVolumeHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/volumes").Handler(PostVolumeHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/volumes").Handler(DeleteVolumeHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/volumes/volume-groups").Handler(GetVolumeGroupMembersHandler(array, logger))
}

func GetVolumesHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumes := []flashclient.Volume{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				volume, err := array.GetVolume(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume by ID", "id", id, "name", volume.Name)
				volumes = append(volumes, *volume)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				volume, err := array.GetVolumeByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume by name", "name", name, "id", volume.Id)
				volumes = append(volumes, volume)
			}
		} else {
			logger.DebugContext(r.Context(), "Getting all volumes")
			volumes = array.GetVolumes()
		}

		logger.DebugContext(r.Context(), "Returning volumes", "count", len(volumes))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		data := flashclient.NewResults(volumes)
		json.NewEncoder(w).Encode(data)
	}
}

func PatchVolumeHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumes := []flashclient.Volume{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				volume, err := array.GetVolume(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume by ID", "id", id, "name", volume.Name)
				volumes = append(volumes, *volume)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				volume, err := array.GetVolumeByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume by name", "name", name, "id", volume.Id)
				volumes = append(volumes, volume)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
		}

		var volumePatch flashclient.VolumePatch
		err := json.NewDecoder(r.Body).Decode(&volumePatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(volumes) > 1 && volumePatch.Name != nil {
			httpJsonError(w, "Cannot rename multiple volumes at once", http.StatusBadRequest)
			return
		}

		updatedVolumes := []flashclient.Volume{}
		for i := range volumes {
			updated, err := array.UpdateVolume(volumes[i].Id, volumePatch)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to update volume", "id", volumes[i].Id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to update volume with ID %s: %v", volumes[i].Id, err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated volume", "id", updated.Id, "name", updated.Name)
			updatedVolumes = append(updatedVolumes, *updated)
		}
		w.Header().Set("Content-Type", "application/json")
		body := flashclient.NewResults(updatedVolumes)
		json.NewEncoder(w).Encode(body)
	}
}

// Create or copy a volume and upsert tags
func PostVolumeHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeNames := splitQueryParam(r.URL.Query().Get("names"))
		if len(volumeNames) < 1 || volumeNames[0] == "" {
			logger.WarnContext(r.Context(), "Missing 'names' query parameter")
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		addToPromotionGroupIds := splitQueryParam(r.URL.Query().Get("add_to_promotion_group_ids"))
		addToPromotionGroupNames := splitQueryParam(r.URL.Query().Get("add_to_promotion_group_names"))
		if len(addToPromotionGroupNames) > 0 && addToPromotionGroupNames[0] != "" {
			for _, name := range addToPromotionGroupNames {
				pg, err := array.GetProtectionGroupByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
					return
				}
				addToPromotionGroupIds = append(addToPromotionGroupIds, pg.Id)
			}
		}
		slices.Sort(addToPromotionGroupIds)
		addToPromotionGroupIds = slices.Compact(addToPromotionGroupIds)

		var volumePost flashclient.VolumePost
		err := json.NewDecoder(r.Body).Decode(&volumePost)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to decode request body", "error", err)
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		volumes := []flashclient.Volume{}
		for _, name := range volumeNames {
			newVolume := fakearray.NewVolumeFromPost(name, volumePost)
			volume, err := array.AddVolume(newVolume)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to add volume", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to add volume: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Added volume", "name", name, "id", volume.Id)
			volumes = append(volumes, *volume)
		}

		w.Header().Set("Content-Type", "application/json")
		result := flashclient.NewResults(volumes)
		json.NewEncoder(w).Encode(result)
	}
}

func DeleteVolumeHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				volume, err := array.GetVolume(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume by ID for deletion", "id", id, "name", volume.Name)
				volumeIds = append(volumeIds, volume.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				volume, err := array.GetVolumeByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume by name for deletion", "name", name, "id", volume.Id)
				volumeIds = append(volumeIds, volume.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, id := range volumeIds {
			err := array.EradicateVolume(id)
			if errors.Is(err, fakearray.ErrVolumeNotFound) {
				logger.WarnContext(r.Context(), "Volume not found during deletion", "id", id)
				httpJsonError(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
				return
			} else if err != nil {
				logger.ErrorContext(r.Context(), "Failed to delete volume", "id", id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to delete volume with ID %s: %v", id, err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Deleted volume", "id", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}
