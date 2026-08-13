package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	faclient "github.com/sanderdescamps/go-purefa"
	"github.com/sanderdescamps/go-purefa-mock/internal/fakearray"
)

func InitConnectionRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/connections").Handler(GetConnectionsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/connections").Handler(PostConnectionsHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/connections").Handler(DeleteConnectionsHandler(array, logger))
}

func GetConnectionsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("host_group_names"))
		hostNames := splitQueryParam(r.URL.Query().Get("host_names"))
		volumeIds := splitQueryParam(r.URL.Query().Get("volume_ids"))
		volumeNames := splitQueryParam(r.URL.Query().Get("volume_names"))
		for _, volumeName := range volumeNames {
			volume, err := array.GetVolumeByName(volumeName)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to get volume", "name", volumeName, "error", err)
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", volumeName), http.StatusNotFound)
				return
			}
			logger.DebugContext(r.Context(), "Resolved volume name to ID", "name", volumeName, "id", volume.Id)
			volumeIds = append(volumeIds, volume.Id)
		}
		slices.Sort(volumeIds)
		volumeIds = slices.Compact(volumeIds)

		filters := []func(*faclient.Connection) bool{}
		if len(hostGroupNames) > 0 {
			filters = append(filters, fakearray.ConnectionsWithHostGroupNames(hostGroupNames...))
		}
		if len(hostNames) > 0 {
			filters = append(filters, fakearray.ConnectionsWithHostNames(hostNames...))
		}
		if len(volumeIds) > 0 {
			filters = append(filters, fakearray.ConnectionsWithVolumeIds(volumeIds...))
		}

		connections := array.GetConnectionsWithFilter(filters...)
		logger.DebugContext(r.Context(), "Returning connections", "count", len(connections))

		w.Header().Set("Content-Type", "application/json")
		data := faclient.NewResults(connections)
		json.NewEncoder(w).Encode(data)
	}
}

func PostConnectionsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("host_group_names"))
		hostNames := splitQueryParam(r.URL.Query().Get("host_names"))
		volumeIds := splitQueryParam(r.URL.Query().Get("volume_ids"))
		volumeNames := splitQueryParam(r.URL.Query().Get("volume_names"))

		if len(hostGroupNames) == 0 && len(hostNames) == 0 {
			httpJsonError(w, "Missing 'host_group_names' or 'host_names' query parameter", http.StatusBadRequest)
			return
		} else if len(hostGroupNames) > 0 && len(hostNames) > 0 {
			httpJsonError(w, "Cannot specify both 'host_group_names' and 'host_names' query parameters", http.StatusBadRequest)
			return
		}

		if len(volumeIds) == 0 && len(volumeNames) == 0 {
			httpJsonError(w, "Missing 'volume_ids' or 'volume_names' query parameter", http.StatusBadRequest)
			return
		} else if len(volumeIds) > 0 && len(volumeNames) > 0 {
			httpJsonError(w, "Cannot specify both 'volume_ids' and 'volume_names' query parameters", http.StatusBadRequest)
			return
		}

		if len(volumeNames) > 0 {
			for _, volumeName := range volumeNames {
				volume, err := array.GetVolumeByName(volumeName)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "name", volumeName, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with name %s not found", volumeName), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Resolved volume name to ID", "name", volumeName, "id", volume.Id)
				volumeIds = append(volumeIds, volume.Id)
			}
		}
		slices.Sort(volumeIds)
		volumeIds = slices.Compact(volumeIds)

		if len(hostGroupNames) > 0 {
			for _, hostGroupName := range hostGroupNames {
				hostGroupMembers, err := array.GetHostGroupMembers([]string{hostGroupName}, nil)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get host group members", "name", hostGroupName, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to get members of host group %s: %v", hostGroupName, err), http.StatusInternalServerError)
					return
				}
				for _, member := range hostGroupMembers {
					hostNames = append(hostNames, member.Member.Name)
				}
				logger.DebugContext(r.Context(), "Resolved host group to host names", "group", hostGroupName, "count", len(hostGroupMembers))
			}
		}
		slices.Sort(hostNames)
		hostNames = slices.Compact(hostNames)

		var postBody faclient.ConnectionPostBody
		err := json.NewDecoder(r.Body).Decode(&postBody)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		connections := []*faclient.Connection{}
		for _, hostName := range hostNames {
			for _, volumeId := range volumeIds {
				connection, err := array.CreateConnection(hostName, volumeId, postBody)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to create connection", "host", hostName, "volume_id", volumeId, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to create connection: %v", err), http.StatusInternalServerError)
					return
				}
				logger.DebugContext(r.Context(), "Created connection", "host", hostName, "volume_id", volumeId)
				connections = append(connections, connection)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		result := faclient.NewResults(connections)
		json.NewEncoder(w).Encode(result)
	}
}

func DeleteConnectionsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := splitQueryParam(r.URL.Query().Get("host_group_names"))
		hostNames := splitQueryParam(r.URL.Query().Get("host_names"))
		volumeIds := splitQueryParam(r.URL.Query().Get("volume_ids"))
		volumeNames := splitQueryParam(r.URL.Query().Get("volume_names"))

		if len(hostGroupNames) == 0 && len(hostNames) == 0 {
			httpJsonError(w, "Missing 'host_group_names' or 'host_names' query parameter", http.StatusBadRequest)
			return
		} else if len(hostGroupNames) > 0 && len(hostNames) > 0 {
			httpJsonError(w, "Cannot specify both 'host_group_names' and 'host_names' query parameters", http.StatusBadRequest)
			return
		}

		if len(volumeIds) == 0 && len(volumeNames) == 0 {
			httpJsonError(w, "Missing 'volume_ids' or 'volume_names' query parameter", http.StatusBadRequest)
			return
		} else if len(volumeIds) > 0 && len(volumeNames) > 0 {
			httpJsonError(w, "Cannot specify both 'volume_ids' and 'volume_names' query parameters", http.StatusBadRequest)
			return
		}

		if len(volumeNames) > 0 {
			for _, volumeName := range volumeNames {
				volume, err := array.GetVolumeByName(volumeName)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get volume", "name", volumeName, "error", err)
					httpJsonError(w, fmt.Sprintf("Volume with name %s not found", volumeName), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Resolved volume name to ID", "name", volumeName, "id", volume.Id)
				volumeIds = append(volumeIds, volume.Id)
			}
		}
		slices.Sort(volumeIds)
		volumeIds = slices.Compact(volumeIds)

		if len(hostGroupNames) > 0 {
			for _, hostGroupName := range hostGroupNames {
				hostGroupMembers, err := array.GetHostGroupMembers([]string{hostGroupName}, nil)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get host group members", "name", hostGroupName, "error", err)
					httpJsonError(w, fmt.Sprintf("Failed to get members of host group %s: %v", hostGroupName, err), http.StatusInternalServerError)
					return
				}
				for _, member := range hostGroupMembers {
					hostNames = append(hostNames, member.Member.Name)
				}
				logger.DebugContext(r.Context(), "Resolved host group to host names", "group", hostGroupName, "count", len(hostGroupMembers))
			}
		}

		for _, hostName := range hostNames {
			for _, volumeId := range volumeIds {
				err := array.DeleteConnection(hostName, volumeId)
				if err != nil {
					if err == fakearray.ErrNotFound {
						logger.WarnContext(r.Context(), "Connection not found during deletion", "host", hostName, "volume_id", volumeId)
						httpJsonError(w, fmt.Sprintf("Connection not found for host %s and volume %s", hostName, volumeId), http.StatusNotFound)
					} else {
						logger.ErrorContext(r.Context(), "Failed to delete connection", "host", hostName, "volume_id", volumeId, "error", err)
						httpJsonError(w, fmt.Sprintf("Failed to delete connection: %v", err), http.StatusInternalServerError)
					}
					return
				}
				logger.DebugContext(r.Context(), "Deleted connection", "host", hostName, "volume_id", volumeId)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
