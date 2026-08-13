package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	faclient "github.com/sanderdescamps/go-purefa"
	"github.com/sanderdescamps/go-purefa-mock/internal/fakearray"
)

func InitProtectionGroupRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/protection-groups").Handler(GetProtectionGroupsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/protection-groups").Handler(PostProtectionGroupHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/protection-groups").Handler(PatchProtectionGroupHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/protection-groups").Handler(DeleteProtectionGroupHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/protection-groups/hosts").Handler(GetProtectionGroupHostMembersHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/protection-groups/host-groups").Handler(GetProtectionGroupHostGroupMembersHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/protection-groups/volumes").Handler(GetProtectionGroupVolumeMembersHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/protection-groups/hosts").Handler(PostProtectionGroupHostMembersHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/protection-groups/host-groups").Handler(PostProtectionGroupHostGroupMembersHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/protection-groups/volumes").Handler(PostProtectionGroupVolumeMembersHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/protection-groups/hosts").Handler(DeleteProtectionGroupHostMembersHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/protection-groups/host-groups").Handler(DeleteProtectionGroupHostGroupMembersHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/protection-groups/volumes").Handler(DeleteProtectionGroupVolumeMembersHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/protection-group-snapshots").Handler(GetProtectionGroupSnapshotsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/protection-group-snapshots").Handler(PostProtectionGroupSnapshotHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/protection-group-snapshots").Handler(PatchProtectionGroupSnapshotHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/protection-group-snapshots").Handler(DeleteProtectionGroupSnapshotHandler(array, logger))
}

func GetProtectionGroupsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		protectionGroups := []faclient.ProtectionGroup{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				pg, err := array.GetProtectionGroup(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found protection group by ID", "id", id, "name", pg.Name)
				protectionGroups = append(protectionGroups, *pg)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				pg, err := array.GetProtectionGroupByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found protection group by name", "name", name, "id", pg.Id)
				protectionGroups = append(protectionGroups, *pg)
			}
		} else {
			logger.DebugContext(r.Context(), "Getting all protection groups")
			protectionGroups = array.GetProtectionGroups()
		}

		logger.DebugContext(r.Context(), "Returning protection groups", "count", len(protectionGroups))
		w.Header().Set("Content-Type", "application/json")
		data := faclient.NewResults(protectionGroups)
		json.NewEncoder(w).Encode(data)
	}
}

func PostProtectionGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		protectionGroupNames := splitQueryParam(r.URL.Query().Get("names"))
		if len(protectionGroupNames) < 1 || protectionGroupNames[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		protectionGroupsCreated := []faclient.ProtectionGroup{}
		for _, name := range protectionGroupNames {
			newProtectionGroup := fakearray.NewProtectionGroupPost(name)
			pg, err := array.AddProtectionGroup(*newProtectionGroup)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to add protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to add protection group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Created protection group", "name", pg.Name, "id", pg.Id)
			protectionGroupsCreated = append(protectionGroupsCreated, *pg)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(faclient.NewResults(protectionGroupsCreated))
	}
}

func PatchProtectionGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		protectionGroupIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				pg, err := array.GetProtectionGroup(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found protection group by ID", "id", id, "name", pg.Name)
				protectionGroupIds = append(protectionGroupIds, pg.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				pg, err := array.GetProtectionGroupByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found protection group by name", "name", name, "id", pg.Id)
				protectionGroupIds = append(protectionGroupIds, pg.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		var protectionGroupPatch faclient.ProtectionGroupPatchBody
		err := json.NewDecoder(r.Body).Decode(&protectionGroupPatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(protectionGroupIds) > 1 && protectionGroupPatch.Name != nil {
			httpJsonError(w, "Cannot specify multiple names in the query parameter when 'name' field is set in the request body", http.StatusBadRequest)
			return
		}

		protectionGroups := []faclient.ProtectionGroup{}
		for _, id := range protectionGroupIds {
			pg, err := array.UpdateProtectionGroup(id, protectionGroupPatch)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to update protection group", "id", id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to update protection group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated protection group", "id", pg.Id, "name", pg.Name)
			protectionGroups = append(protectionGroups, *pg)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(faclient.NewResults(protectionGroups))
	}
}

func DeleteProtectionGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		protectionGroupIds := []string{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				pg, err := array.GetProtectionGroup(id)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "id", id, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with ID %s not found", id), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found protection group by ID for deletion", "id", id, "name", pg.Name)
				protectionGroupIds = append(protectionGroupIds, pg.Id)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				pg, err := array.GetProtectionGroupByName(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found protection group by name for deletion", "name", name, "id", pg.Id)
				protectionGroupIds = append(protectionGroupIds, pg.Id)
			}
		} else {
			httpJsonError(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, id := range protectionGroupIds {
			err := array.EradicateProtectionGroup(id)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to delete protection group", "id", id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to delete protection group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Deleted protection group", "id", id)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetProtectionGroupHostMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		slices.Sort(memberNames)
		memberNames = slices.Compact(memberNames)

		members, err := array.GetProtectionGroupHostMembers(groupIds, memberNames)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get protection group host members", "groups", groupIds, "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get protection group host members for group IDs %v and member names %v", groupIds, memberNames), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Returning protection group host members", "count", len(members))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(members)
		json.NewEncoder(w).Encode(data)
	}
}

func GetProtectionGroupHostGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		slices.Sort(memberNames)
		memberNames = slices.Compact(memberNames)

		members, err := array.GetProtectionGroupHostGroupMembers(groupIds, memberNames)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get protection group host group members", "groups", groupIds, "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get protection group host group members for group IDs %v and member names %v", groupIds, memberNames), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Returning protection group host group members", "count", len(members))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(members)
		json.NewEncoder(w).Encode(data)
	}
}

func GetProtectionGroupVolumeMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberIds := splitQueryParam(r.URL.Query().Get("member_ids"))
		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		for _, name := range memberNames {
			volume, err := array.GetVolumeByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get volume", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved volume member name to ID", "name", name, "id", volume.Id)
			memberIds = append(memberIds, volume.Id)
		}
		slices.Sort(memberIds)
		memberIds = slices.Compact(memberIds)

		members, err := array.GetProtectionGroupVolumeMembers(groupIds, memberIds)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get protection group volume members", "groups", groupIds, "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get protection group volume members for group IDs %v and member IDs %v", groupIds, memberIds), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Returning protection group volume members", "count", len(members))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(members)
		json.NewEncoder(w).Encode(data)
	}
}

func PostProtectionGroupHostMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		slices.Sort(memberNames)
		memberNames = slices.Compact(memberNames)

		members := []faclient.ProtectionGroupHostMember{}
		for _, groupId := range groupIds {
			for _, memberName := range memberNames {
				member, err := array.AddProtectionGroupHostMember(groupId, memberName)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to add protection group host member", "group_id", groupId, "member", memberName, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to add protection group host member for group ID %s and member name %s: %v", groupId, memberName, err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Added protection group host member", "group_id", groupId, "member", memberName)
				members = append(members, *member)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(members)
		json.NewEncoder(w).Encode(data)
	}
}

func PostProtectionGroupHostGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		slices.Sort(memberNames)
		memberNames = slices.Compact(memberNames)

		members := []faclient.ProtectionGroupHostGroupMember{}
		for _, groupId := range groupIds {
			for _, memberName := range memberNames {
				member, err := array.AddProtectionGroupHostGroupMember(groupId, memberName)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to add protection group host group member", "group_id", groupId, "member", memberName, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to add protection group host group member for group ID %s and member name %s: %v", groupId, memberName, err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Added protection group host group member", "group_id", groupId, "member", memberName)
				members = append(members, *member)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(members)
		json.NewEncoder(w).Encode(data)
	}
}

func PostProtectionGroupVolumeMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberIds := splitQueryParam(r.URL.Query().Get("member_ids"))
		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		for _, name := range memberNames {
			volume, err := array.GetVolumeByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get volume", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved volume member name to ID", "name", name, "id", volume.Id)
			memberIds = append(memberIds, volume.Id)
		}
		slices.Sort(memberIds)
		memberIds = slices.Compact(memberIds)

		members := []faclient.ProtectionGroupVolumeMember{}
		for _, groupId := range groupIds {
			for _, memberId := range memberIds {
				member, err := array.AddProtectionGroupVolumeMember(groupId, memberId)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to add protection group volume member", "group_id", groupId, "member_id", memberId, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to add protection group volume member for group ID %s and member ID %s: %v", groupId, memberId, err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Added protection group volume member", "group_id", groupId, "member_id", memberId)
				members = append(members, *member)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(members)
		json.NewEncoder(w).Encode(data)
	}
}

func DeleteProtectionGroupHostMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		slices.Sort(memberNames)
		memberNames = slices.Compact(memberNames)

		for _, groupId := range groupIds {
			for _, memberName := range memberNames {
				err := array.RemoveProtectionGroupHostMember(groupId, memberName)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to remove protection group host member", "group_id", groupId, "member", memberName, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to remove protection group host member for group ID %s and member name %s: %v", groupId, memberName, err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Removed protection group host member", "group_id", groupId, "member", memberName)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func DeleteProtectionGroupHostGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		slices.Sort(memberNames)
		memberNames = slices.Compact(memberNames)

		for _, groupId := range groupIds {
			for _, memberName := range memberNames {
				err := array.RemoveProtectionGroupHostGroupMember(groupId, memberName)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to remove protection group host group member", "group_id", groupId, "member", memberName, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to remove protection group host group member for group ID %s and member name %s: %v", groupId, memberName, err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Removed protection group host group member", "group_id", groupId, "member", memberName)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func DeleteProtectionGroupVolumeMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIds := splitQueryParam(r.URL.Query().Get("group_ids"))
		groupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		for _, name := range groupNames {
			group, err := array.GetProtectionGroupByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group name to ID", "name", name, "id", group.Id)
			groupIds = append(groupIds, group.Id)
		}
		slices.Sort(groupIds)
		groupIds = slices.Compact(groupIds)

		memberIds := splitQueryParam(r.URL.Query().Get("member_ids"))
		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		for _, name := range memberNames {
			volume, err := array.GetVolumeByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get volume", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved volume member name to ID", "name", name, "id", volume.Id)
			memberIds = append(memberIds, volume.Id)
		}
		slices.Sort(memberIds)
		memberIds = slices.Compact(memberIds)

		for _, groupId := range groupIds {
			for _, memberId := range memberIds {
				err := array.RemoveProtectionGroupVolumeMember(groupId, memberId)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to remove protection group volume member", "group_id", groupId, "member_id", memberId, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to remove protection group volume member for group ID %s and member ID %s: %v", groupId, memberId, err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Removed protection group volume member", "group_id", groupId, "member_id", memberId)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func GetProtectionGroupSnapshotsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := splitQueryParam(r.URL.Query().Get("ids"))
		names := splitQueryParam(r.URL.Query().Get("names"))
		sourceIds := splitQueryParam(r.URL.Query().Get("source_ids"))
		sourceNames := splitQueryParam(r.URL.Query().Get("source_names"))

		expectedResults := len(ids) + len(names) + len(sourceIds) + len(sourceNames)
		logger.DebugContext(r.Context(), "Getting protection group snapshots", "expected", expectedResults)

		filters := []func(*faclient.ProtectionGroupSnapshot) bool{}
		if len(ids) > 0 {
			filters = append(filters, fakearray.WithIDs[faclient.ProtectionGroupSnapshot](ids...))
		}
		if len(names) > 0 {
			filters = append(filters, fakearray.WithNames[faclient.ProtectionGroupSnapshot](names...))
		}
		if len(sourceIds) > 0 {
			filters = append(filters, fakearray.ProtectionGroupSnapshotsWithSourceIds(sourceIds...))
		}
		if len(sourceNames) > 0 {
			filters = append(filters, fakearray.ProtectionGroupSnapshotsWithSourceNames(sourceNames...))
		}

		snapshots, err := array.GetProtectionGroupSnapshots(filters...)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get protection group snapshots", "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get protection group snapshots: %v", err), http.StatusInternalServerError)
			return
		} else if expectedResults > 0 && len(snapshots) != expectedResults {
			logger.WarnContext(r.Context(), "Unexpected number of protection group snapshots", "expected", expectedResults, "found", len(snapshots))
			httpJsonError(w, fmt.Sprintf("Expected %d protection group snapshots but found %d", expectedResults, len(snapshots)), http.StatusNotFound)
			return
		}

		logger.DebugContext(r.Context(), "Returning protection group snapshots", "count", len(snapshots))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(snapshots)
		json.NewEncoder(w).Encode(data)
	}
}

func PostProtectionGroupSnapshotHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourceIds := splitQueryParam(r.URL.Query().Get("source_ids"))
		sourceNames := splitQueryParam(r.URL.Query().Get("source_names"))
		for _, sourceName := range sourceNames {
			source, err := array.GetProtectionGroupByName(sourceName)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group", "name", sourceName, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with name %s not found", sourceName), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group source name to ID", "name", sourceName, "id", source.Id)
			sourceIds = append(sourceIds, source.Id)
		}
		slices.Sort(sourceIds)
		sourceIds = slices.Compact(sourceIds)

		body := faclient.ProtectionGroupSnapshotPostBody{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		snapshots := []faclient.ProtectionGroupSnapshot{}
		for _, sourceId := range sourceIds {
			snapshot, err := array.CreateProtectionGroupSnapshot(sourceId, body)
			if errors.Is(err, fakearray.ErrNotFound) {
				logger.ErrorContext(r.Context(), "Protection group source not found", "source_id", sourceId, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group with source ID %s not found", sourceId), http.StatusNotFound)
				return
			} else if err != nil {
				logger.ErrorContext(r.Context(), "Failed to create protection group snapshot", "source_id", sourceId, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to create protection group snapshot for source ID %s: %v", sourceId, err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Created protection group snapshot", "source_id", sourceId, "id", snapshot.Id)
			snapshots = append(snapshots, *snapshot)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(snapshots)
		json.NewEncoder(w).Encode(data)
	}
}

func PatchProtectionGroupSnapshotHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := splitQueryParam(r.URL.Query().Get("ids"))
		names := splitQueryParam(r.URL.Query().Get("names"))

		filters := []func(*faclient.ProtectionGroupSnapshot) bool{}
		if len(ids) > 0 {
			filters = append(filters, fakearray.WithIDs[faclient.ProtectionGroupSnapshot](ids...))
		}
		if len(names) > 0 {
			filters = append(filters, fakearray.WithNames[faclient.ProtectionGroupSnapshot](names...))
		}

		snapshots, err := array.GetProtectionGroupSnapshots(filters...)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get protection group snapshots", "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get protection group snapshots: %v", err), http.StatusInternalServerError)
			return
		}

		body := faclient.ProtectionGroupSnapshotPatchBody{}
		err = json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		updatedSnapshots := []faclient.ProtectionGroupSnapshot{}
		for _, snapshot := range snapshots {
			updatedSnapshot, err := array.UpdateProtectionGroupSnapshot(snapshot.Id, body)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to update protection group snapshot", "id", snapshot.Id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to update protection group snapshot with ID %s: %v", snapshot.Id, err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated protection group snapshot", "id", updatedSnapshot.Id)
			updatedSnapshots = append(updatedSnapshots, *updatedSnapshot)
		}

		logger.DebugContext(r.Context(), "Returning updated protection group snapshots", "count", len(updatedSnapshots))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.NewResults(updatedSnapshots)
		json.NewEncoder(w).Encode(data)
	}
}

func DeleteProtectionGroupSnapshotHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := splitQueryParam(r.URL.Query().Get("ids"))
		names := splitQueryParam(r.URL.Query().Get("names"))
		for _, name := range names {
			snapshot, err := array.GetProtectionGroupSnapshotByName(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get protection group snapshot", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Protection group snapshot with name %s not found", name), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved protection group snapshot name to ID", "name", name, "id", snapshot.Id)
			ids = append(ids, snapshot.Id)
		}
		slices.Sort(ids)
		ids = slices.Compact(ids)

		for _, id := range ids {
			err := array.EradicateProtectionGroupSnapshot(id)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to delete protection group snapshot", "id", id, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to eradicate protection group snapshot with ID %s: %v", id, err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Deleted protection group snapshot", "id", id)
		}
		w.WriteHeader(http.StatusOK)
	}
}
