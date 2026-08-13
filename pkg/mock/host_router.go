package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sanderdescamps/go-purefa-mock/internal/fakearray"
	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func InitHostRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/hosts").Handler(GetHostsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/hosts").Handler(PostHostHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/hosts").Handler(PatchHostHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/hosts").Handler(DeleteHostHandler(array, logger))
}

func GetHostsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hosts := []flashclient.Host{}
		if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				host, err := array.GetHost(name)
				if err != nil {
					logger.ErrorContext(r.Context(), "Failed to get host", "name", name, "error", err)
					httpJsonError(w, fmt.Sprintf("Host with name %s not found", name), http.StatusNotFound)
					return
				}
				logger.DebugContext(r.Context(), "Found host by name", "name", name)
				hosts = append(hosts, *host)
			}
		} else {
			logger.DebugContext(r.Context(), "Getting all hosts")
			hosts = array.GetHosts()
		}

		logger.DebugContext(r.Context(), "Returning hosts", "count", len(hosts))
		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(hosts)
		json.NewEncoder(w).Encode(data)
	}
}

func PostHostHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostNames := splitQueryParam(r.URL.Query().Get("names"))
		if len(hostNames) < 1 || hostNames[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var hostPost flashclient.HostPostBody
		err := json.NewDecoder(r.Body).Decode(&hostPost)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		hostsCreated := []flashclient.Host{}
		for _, name := range hostNames {
			newHost := fakearray.NewHostPost(name, hostPost)
			host, err := array.AddHost(newHost)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to add host", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to add host: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Created host", "name", host.Name)
			hostsCreated = append(hostsCreated, *host)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(hostsCreated))
	}
}

func PatchHostHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := splitQueryParam(r.URL.Query().Get("names"))
		if len(names) == 0 || names[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var hostPatch flashclient.HostPatchBody
		err := json.NewDecoder(r.Body).Decode(&hostPatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(names) > 1 && hostPatch.Name != nil && *hostPatch.Name != "" {
			httpJsonError(w, "Cannot specify multiple names in the query parameter when 'name' field is set in the request body", http.StatusBadRequest)
			return
		}

		hosts := []flashclient.Host{}
		for _, name := range names {
			host, err := array.UpdateHost(name, hostPatch)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to update host", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to update host: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Updated host", "name", host.Name)
			hosts = append(hosts, *host)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flashclient.NewResults(hosts))
	}
}

func DeleteHostHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := splitQueryParam(r.URL.Query().Get("names"))
		if len(names) < 1 || names[0] == "" {
			httpJsonError(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, name := range names {
			err := array.DeleteHost(name)
			if err != nil {
				logger.ErrorContext(r.Context(), "Failed to delete host", "name", name, "error", err)
				httpJsonError(w, fmt.Sprintf("Failed to delete host: %v", err), http.StatusInternalServerError)
				return
			}
			logger.DebugContext(r.Context(), "Deleted host", "name", name)
		}

		w.WriteHeader(http.StatusOK)
	}
}
