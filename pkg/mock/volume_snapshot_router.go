package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func InitVolumeSnapshotRouter(r *mux.Router, array *fakearray.Array, logger *slog.Logger) {
	r.Methods("GET").Path("/api/{api_version}/volume-snapshots").Handler(GetVolumeSnapshotsHandler(array, logger))
	r.Methods("POST").Path("/api/{api_version}/volume-snapshots").Handler(PostVolumeSnapshotHandler(array, logger))
	r.Methods("PATCH").Path("/api/{api_version}/volume-snapshots").Handler(PatchVolumeSnapshotHandler(array, logger))
	r.Methods("DELETE").Path("/api/{api_version}/volume-snapshots").Handler(DeleteVolumeSnapshotHandler(array, logger))

}

func GetVolumeSnapshotsHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourceIds := splitQueryParam(r.URL.Query().Get("source_ids"))
		for _, name := range splitQueryParam(r.URL.Query().Get("source_names")) {
			volume, err := array.GetVolumeByName(name)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
				return
			}
			sourceIds = append(sourceIds, volume.Id)
		}
		slices.Sort(sourceIds)
		sourceIds = slices.Compact(sourceIds)

		snapshots := []flashclient.VolumeSnapshot{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range splitQueryParam(ids) {
				snapshot, err := array.GetVolumeSnapshot(id)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume snapshot with ID %s not found", id), http.StatusNotFound)
					return
				}
				snapshots = append(snapshots, *snapshot)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range splitQueryParam(names) {
				snapshot, err := array.GetVolumeSnapshotByName(name)
				if err != nil {
					httpJsonError(w, fmt.Sprintf("Volume snapshot with name %s not found", name), http.StatusNotFound)
					return
				}
				snapshots = append(snapshots, *snapshot)
			}
		} else if len(sourceIds) > 0 {
			snaps, err := array.GetVolumeSnapshotsForSources(sourceIds...)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to get volume snapshots for source IDs %v: %v", sourceIds, err), http.StatusNotFound)
				return
			}
			snapshots = append(snapshots, snaps...)
		} else {
			snapshots = array.GetVolumeSnapshots()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults(snapshots)
		json.NewEncoder(w).Encode(data)
	}
}

// Create a volume snapshot
func PostVolumeSnapshotHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourceIds := splitQueryParam(r.URL.Query().Get("source_ids"))
		for _, name := range splitQueryParam(r.URL.Query().Get("source_names")) {
			volume, err := array.GetVolumeByName(name)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
				return
			}
			sourceIds = append(sourceIds, volume.Id)
		}
		slices.Sort(sourceIds)
		sourceIds = slices.Compact(sourceIds)

		volumeSnapPost := flashclient.VolumeSnapshotPostBody{}
		err := json.NewDecoder(r.Body).Decode(&volumeSnapPost)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		snapshots := []flashclient.VolumeSnapshot{}
		for _, sourceId := range sourceIds {
			snapshot, err := array.CreateVolumeSnapshot(sourceId, volumeSnapPost)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to create volume snapshot for source ID %s: %v", sourceId, err), http.StatusInternalServerError)
				return
			}
			snapshots = append(snapshots, *snapshot)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults(snapshots)
		json.NewEncoder(w).Encode(data)
	}
}

func PatchVolumeSnapshotHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshotIds := splitQueryParam(r.URL.Query().Get("ids"))
		snapshotNames := splitQueryParam(r.URL.Query().Get("names"))
		for _, name := range snapshotNames {
			snapshot, err := array.GetVolumeSnapshotByName(name)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Volume snapshot with name %s not found", name), http.StatusNotFound)
				return
			}
			snapshotIds = append(snapshotIds, snapshot.Id)
		}
		slices.Sort(snapshotIds)
		snapshotIds = slices.Compact(snapshotIds)

		var volumeSnapPatch flashclient.VolumeSnapshotPatchBody
		err := json.NewDecoder(r.Body).Decode(&volumeSnapPatch)
		if err != nil {
			httpJsonError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		snapshots := []flashclient.VolumeSnapshot{}
		for _, snapshotId := range snapshotIds {
			snapshot, err := array.UpdateVolumeSnapshot(snapshotId, volumeSnapPatch)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to update volume snapshot with ID %s: %v", snapshotId, err), http.StatusInternalServerError)
				return
			}
			snapshots = append(snapshots, *snapshot)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := flashclient.NewResults(snapshots)
		json.NewEncoder(w).Encode(data)
	}
}

func DeleteVolumeSnapshotHandler(array *fakearray.Array, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshotIds := splitQueryParam(r.URL.Query().Get("ids"))
		snapshotNames := splitQueryParam(r.URL.Query().Get("names"))
		for _, name := range snapshotNames {
			snapshot, err := array.GetVolumeSnapshotByName(name)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Volume snapshot with name %s not found", name), http.StatusNotFound)
				return
			}
			snapshotIds = append(snapshotIds, snapshot.Id)
		}
		slices.Sort(snapshotIds)
		snapshotIds = slices.Compact(snapshotIds)

		for _, snapshotId := range snapshotIds {
			err := array.EradicateVolumeSnapshot(snapshotId)
			if err != nil {
				httpJsonError(w, fmt.Sprintf("Failed to delete volume snapshot with ID %s: %v", snapshotId, err), http.StatusBadRequest)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
