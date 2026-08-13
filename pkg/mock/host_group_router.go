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

func InitHostGroupRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/host-groups").Handler(GetHostGroupsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/host-groups").Handler(PostHostGroupHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/host-groups").Handler(PatchHostGroupHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/host-groups").Handler(DeleteHostGroupHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/host-groups/hosts").Handler(GetHostGroupMembersHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/host-groups/hosts").Handler(PostHostGroupMembersHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/host-groups/hosts").Handler(DeleteHostGroupMembersHandler(array, logger))
}

func GetHostGroupsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("names"))
		hostGroups := []flashclient.HostGroup{}
		if len(hostGroupNames) > 0 {
			for _, name := range hostGroupNames {
				hostGroup, err := array.GetHostGroup(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get host group", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Host group with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found host group by name", "name", name)
				hostGroups = append(hostGroups, *hostGroup)
			}
		} else {
			logger.DebugContext(r.Context(), "Getting all host groups")
			hostGroups = array.GetHostGroups()
		}

		logger.DebugContext(r.Context(), "Returning host groups", "count", len(hostGroups))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults(hostGroups)
		json.NewEncoder(w).Encode(data)
	}
}

func PostHostGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("names"))
		if len(hostGroupNames) < 1 || hostGroupNames[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		hostGroupsCreated := []flashclient.HostGroup{}
		for _, name := range hostGroupNames {
			newHostGroup := fakearray.NewHostGroup(name)
			hostGroup, err := array.AddHostGroup(*newHostGroup)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to add host group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to add host group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Created host group", "name", hostGroup.Name)
			hostGroupsCreated = append(hostGroupsCreated, *hostGroup)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(hostGroupsCreated))
	}
}

func PatchHostGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := splitQueryParam(r.URL.Query().Get("names"))
		if len(names) < 1 || names[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var hostGroupPatch flashclient.HostGroupPatchBody
		err := json.NewDecoder(r.Body).Decode(&hostGroupPatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(names) > 1 && hostGroupPatch.Name != "" {
			httpJsonError(w, "Cannot specify multiple names in the query parameter when 'name' field is set in the request body", http.StatusBadRequest)
			return
		}

		hostGroups := []flashclient.HostGroup{}
		for _, name := range names {
			hostGroup, err := array.UpdateHostGroup(name, hostGroupPatch)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to update host group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to update host group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated host group", "name", hostGroup.Name)
			hostGroups = append(hostGroups, *hostGroup)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(hostGroups))
	}
}

func DeleteHostGroupHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := splitQueryParam(r.URL.Query().Get("names"))
		if len(names) == 0 || names[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, name := range names {
			err := array.DeleteHostGroup(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to delete host group", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to delete host group: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Deleted host group", "name", name)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetHostGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))

		logger.DebugContext(r.Context(), "Getting host group members", "groups", hostGroupNames, "members", memberNames)
		members, err := array.GetHostGroupMembers(hostGroupNames, memberNames)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get host group members", "groups", hostGroupNames, "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get host group members: %v", err), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Returning host group members", "count", len(members))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(members))
	}
}

func PostHostGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		if len(hostGroupNames) < 1 || hostGroupNames[0] == "" {
			httpJsonError(w, "Missing 'group_names' query parameter", http.StatusBadRequest)
			return
		} else if len(hostGroupNames) > 1 {
			httpJsonError(w, "Cannot specify multiple group names in the query parameter", http.StatusBadRequest)
			return
		}

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		if len(memberNames) < 1 || memberNames[0] == "" {
			httpJsonError(w, "Missing 'member_names' query parameter", http.StatusBadRequest)
			return
		}

		members, err := array.AddHostGroupMembers(hostGroupNames[0], memberNames)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to add host group members", "group", hostGroupNames[0], "members", memberNames, "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to add host group members: %v", err), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Added host group members", "group", hostGroupNames[0], "count", len(members))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(members))
	}
}

func DeleteHostGroupMembersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("group_names"))
		if len(hostGroupNames) < 1 || hostGroupNames[0] == "" {
			httpJsonError(w, "Missing 'group_names' query parameter", http.StatusBadRequest)
			return
		} else if len(hostGroupNames) > 1 {
			httpJsonError(w, "Cannot specify multiple group names in the query parameter", http.StatusBadRequest)
			return
		}

		memberNames := splitQueryParam(r.URL.Query().Get("member_names"))
		if len(memberNames) < 1 || memberNames[0] == "" {
			httpJsonError(w, "Missing 'member_names' query parameter", http.StatusBadRequest)
			return
		}

		err := array.DeleteHostGroupMembers(hostGroupNames[0], memberNames)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to delete host group members", "group", hostGroupNames[0], "members", memberNames, "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to delete host group members: %v", err), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(r.Context(), "Deleted host group members", "group", hostGroupNames[0], "members", memberNames)

		w.WriteHeader(http.StatusOK)
	}
}
