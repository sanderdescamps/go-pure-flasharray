package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func InitMiscRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/alerts").Handler(GetAlertsHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/arrays").Handler(GetArraysHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/controllers").Handler(GetControllersHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/drives").Handler(GetDrivesHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/hardware").Handler(GetHardwareHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/network-interfaces").Handler(GetNetworkInterfacesHandler(array, logger))
	r.Methods("GET").Path("/api/{api_version}/ports").Handler(GetPortsHandler(array, logger))
}

func GetAlertsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := []func(*flashclient.Alert) bool{}
		names := splitQueryParam(r.URL.Query().Get("names"))
		if len(names) > 0 {
			filters = append(filters, fakearray.WithNames[flashclient.Alert](names...))
		}
		ids := splitQueryParam(r.URL.Query().Get("ids"))
		if len(ids) > 0 {
			filters = append(filters, fakearray.WithIDs[flashclient.Alert](ids...))
		}
		flagged := r.URL.Query().Get("flagged")
		if flagged != "" {
			flaggedBool, err := strconv.ParseBool(flagged)
			if err == nil {
				filters = append(filters, fakearray.AlertsWithFlagged(flaggedBool))
			}
		}

		alerts, err := array.GetAlerts(filters...)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get alerts", "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get alerts: %v", err), http.StatusInternalServerError)
			return
		}

		logger.DebugContext(r.Context(), "Returning alerts", "count", len(alerts))
		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(alerts)
		json.NewEncoder(w).Encode(data)
	}
}

func GetArraysHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		arrays, err := array.GetArrays()
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to get arrays", "error", err)
			httpJsonError(w, fmt.Sprintf("Failed to get arrays: %v", err), http.StatusInternalServerError)
			return
		}

		logger.DebugContext(r.Context(), "Returning arrays", "count", len(arrays))
		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(arrays)
		json.NewEncoder(w).Encode(data)
	}
}

func GetControllersHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controllers := array.GetControllers()
		logger.DebugContext(r.Context(), "Returning controllers", "count", len(controllers))

		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(controllers)
		json.NewEncoder(w).Encode(data)
	}
}

func GetDrivesHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := splitQueryParam(r.URL.Query().Get("names"))
		filters := []func(*flashclient.Drive) bool{}
		if len(names) > 0 {
			filters = append(filters, fakearray.WithNames[flashclient.Drive](names...))
		}
		drives := array.GetDrives(filters...)
		logger.DebugContext(r.Context(), "Returning drives", "count", len(drives))

		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(drives)
		json.NewEncoder(w).Encode(data)
	}
}

func GetHardwareHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := splitQueryParam(r.URL.Query().Get("names"))
		filters := []func(*flashclient.Hardware) bool{}
		if len(names) > 0 {
			filters = append(filters, fakearray.WithNames[flashclient.Hardware](names...))
		}
		hardware := array.GetHardware(filters...)
		logger.DebugContext(r.Context(), "Returning hardware", "count", len(hardware))

		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(hardware)
		json.NewEncoder(w).Encode(data)
	}
}

func GetNetworkInterfacesHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		networkInterfaces := array.GetNetworkInterfaces()
		logger.DebugContext(r.Context(), "Returning network interfaces", "count", len(networkInterfaces))

		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(networkInterfaces)
		json.NewEncoder(w).Encode(data)
	}
}

func GetPortsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ports := array.GetPorts()
		logger.DebugContext(r.Context(), "Returning ports", "count", len(ports))

		w.Header().Set("Content-Type", "application/json")
		data := flashclient.NewResults(ports)
		json.NewEncoder(w).Encode(data)
	}
}
