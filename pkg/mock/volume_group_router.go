package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/sanderdescamps/go-purefa-mock/internal/fakearray"
	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func InitVolumeGroupRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/volume-groups").Handler(GetVolumeGroupsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/volume-groups").Handler(PostVolumeGroupHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/volume-groups").Handler(PatchVolumeGroupHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/volume-groups").Handler(DeleteVolumeGroupHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/volume-groups/volume").Handler(GetVolumeGroupMembersHandler(array, logger))
}

func GetVolumeGroupsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vgroups := []flashclient.VolumeGroup{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				vgroup, err := array.GetVolumeGroup(id)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume group with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume group by ID", "id", id, "name", vgroup.Name)
				vgroups = append(vgroups, *vgroup)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				vgroup, err := array.GetVolumeGroupByName(name)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume group by name", "name", name, "id", vgroup.Id)
				vgroups = append(vgroups, *vgroup)
			}
		} else {
			logger.DebugContext(r.Context(), "Getting all volume groups")
			vgroups = array.GetVolumeGroups()
		}

		logger.DebugContext(r.Context(), "Returning volume groups", "count", len(vgroups))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults(vgroups)
		json.NewEncoder(w).Encode(data)
	}
}

// Create or copy a volume group and upsert tags
func PostVolumeGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeNames := splitQueryParam(r.URL.Query().Get("names"))
		if len(volumeNames) < 1 || volumeNames[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var volumeGroupPost flashclient.VolumeGroupPost
		err := json.NewDecoder(r.Body).Decode(&volumeGroupPost)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		volumeGroupResult := flashclient.NewResults([]flashclient.VolumeGroup{})
		for _, name := range volumeNames {
			vgroup, err := array.CreateVolumeGroup(name, volumeGroupPost)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to add volume group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Created volume group", "name", name, "id", vgroup.Id)
			volumeGroupResult.Items = append(volumeGroupResult.Items, *vgroup)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(volumeGroupResult)
	}
}

func PatchVolumeGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeGroupIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				vgroup, err := array.GetVolumeGroup(id)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume group with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume group by ID", "id", id, "name", vgroup.Name)
				volumeGroupIds = append(volumeGroupIds, vgroup.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				vgroup, err := array.GetVolumeGroupByName(name)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume group by name", "name", name, "id", vgroup.Id)
				volumeGroupIds = append(volumeGroupIds, vgroup.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		var volumeGroupPatch flashclient.VolumeGroupPatch
		err := json.NewDecoder(r.Body).Decode(&volumeGroupPatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(volumeGroupIds) > 1 && volumeGroupPatch.Name != nil {
			httpJsonError(w, "Cannot rename multiple volume groups at once", http.StatusBadRequest)
			return

		}

		volumeGroups := []flashclient.VolumeGroup{}
		for _, id := range volumeGroupIds {
			updatedVgroup, err := array.UpdateVolumeGroup(id, volumeGroupPatch)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to update volume group with ID %s: %v", id, err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated volume group", "id", updatedVgroup.Id, "name", updatedVgroup.Name)
			volumeGroups = append(volumeGroups, *updatedVgroup)
		}
		w.Header().Set("Content-Type", "application/json")
		body := flashclient.NewResults(volumeGroups)
		json.NewEncoder(w).Encode(body)
	}
}

func DeleteVolumeGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeGroupIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				vgroup, err := array.GetVolumeGroup(id)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume group with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume group by ID for deletion", "id", id, "name", vgroup.Name)
				volumeGroupIds = append(volumeGroupIds, vgroup.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				vgroup, err := array.GetVolumeGroupByName(name)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found volume group by name for deletion", "name", name, "id", vgroup.Id)
				volumeGroupIds = append(volumeGroupIds, vgroup.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, id := range volumeGroupIds {
			err := array.EradicateVolumeGroup(id)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to delete volume group with ID %s: %v", id, err), http.StatusBadRequest)
				return
			}
			logger.DebugContext(r.Context(), "Deleted volume group", "id", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetVolumeGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetVolumeGroupByName(name)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved volume group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberIds := splitQueryParam(r.URL.Query().Get("member_ids"))
		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		for _, name := range memberNames {
			volume, err := array.GetVolumeByName(name)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved volume member name to ID", "name", name, "id", volume.Id)
			memberIds = append(memberIds, volume.Id)
		}
		slices.Sort(memberIds)
		memberIds = slices.Compact(memberIds)

		arrayMembers, err := array.GetVolumeGroupMembers(groupIds, memberIds)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Failed to get volume group members: %v", err), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Returning volume group members", "count", len(arrayMembers))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults(arrayMembers)
		json.NewEncoder(w).Encode(data)
	}
}
