package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	test_data "github.com/sanderdescamps/go-purefa-mock/mock/test-data"

	"github.com/gorilla/mux"
	faclient "github.com/sanderdescamps/go-purefa"
)

type Mock struct {
	*http.Server
	APIToken        string
	Username        string
	sessionTokenKey string
	DisableAuth     bool
}

func NewMock(array *Array) *Mock {
	m := &Mock{
		Username:    "fake-auth-user",
		APIToken:    "fake-auth-token",
		DisableAuth: false,
	}

	router := mux.NewRouter()
	logger := log.New(os.Stdout, "mock: ", log.Ltime)
	router.Use(NewLogMiddleware(logger).Func())

	router.Methods("GET").Path("/api/api_version").HandlerFunc(GetVersionHandler(array))

	versionRouter := router.NewRoute().Subrouter()
	versionRouter.Use(m.versionMiddleware(array))
	versionRouter.Methods("POST").Path("/api/{api_version}/login").HandlerFunc(m.loginHandler)

	authRouter := versionRouter.NewRoute().Subrouter()
	authRouter.Use(m.authMiddleware, m.versionMiddleware(array))
	authRouter.Methods("GET").Path("/api/{api_version}/volumes").Handler(GetVolumesHandler(array))
	authRouter.Methods("PATCH").Path("/api/{api_version}/volumes").Handler(PatchVolumeHandler(array))
	authRouter.Methods("POST").Path("/api/{api_version}/volumes").Handler(PostVolumeHandler(array))
	authRouter.Methods("DELETE").Path("/api/{api_version}/volumes").Handler(DeleteVolumeHandler(array))
	authRouter.Methods("GET").Path("/api/{api_version}/volume-groups").Handler(GetVolumeGroupsHandler(array))
	authRouter.Methods("POST").Path("/api/{api_version}/volume-groups").Handler(PostVolumeGroupHandler(array))
	authRouter.Methods("PATCH").Path("/api/{api_version}/volume-groups").Handler(PatchVolumeGroupHandler(array))
	authRouter.Methods("DELETE").Path("/api/{api_version}/volume-groups").Handler(DeleteVolumeGroupHandler(array))
	authRouter.Methods("GET").Path("/api/{api_version}/hosts").Handler(GetHostsHandler(array))
	authRouter.Methods("POST").Path("/api/{api_version}/hosts").Handler(PostHostHandler(array))
	authRouter.Methods("PATCH").Path("/api/{api_version}/hosts").Handler(PatchHostHandler(array))
	authRouter.Methods("DELETE").Path("/api/{api_version}/hosts").Handler(DeleteHostHandler(array))
	authRouter.Methods("GET").Path("/api/{api_version}/host-groups").Handler(GetHostGroupsHandler(array))
	authRouter.Methods("POST").Path("/api/{api_version}/host-groups").Handler(PostHostGroupHandler(array))
	authRouter.Methods("PATCH").Path("/api/{api_version}/host-groups").Handler(PatchHostGroupHandler(array))
	authRouter.Methods("DELETE").Path("/api/{api_version}/host-groups").Handler(DeleteHostGroupHandler(array))

	m.Server = &http.Server{
		Handler: router,
	}
	return m
}

func NewMockWithTestData(apiToken string) (*Mock, error) {
	array, err := NewArrayFromFS(test_data.TestData, ".")
	if err != nil {
		return nil, err
	}
	mock := NewMock(array)
	mock.APIToken = apiToken
	return mock, nil
}

func (m *Mock) GenerateSessionToken() string {
	if m.sessionTokenKey == "" {
		m.sessionTokenKey = NewRandomString(32)
	}
	return m.sessionTokenKey
}

func (m *Mock) Start(addr string, port int) error {
	m.Addr = fmt.Sprintf("%s:%d", addr, port)
	fmt.Printf("Starting mock server on %s\n", m.Addr)
	if err := m.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (m *Mock) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.Shutdown(ctx)
}

type LogMiddleware struct {
	logger *log.Logger
}

func NewLogMiddleware(logger *log.Logger) *LogMiddleware {
	return &LogMiddleware{logger: logger}
}

func (m *LogMiddleware) Func() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.logger.Printf(
				"%s %s %s",
				r.Method,
				r.URL.Path,
				r.URL.Query().Encode(),
			)
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Mock) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if xAuthToken := r.Header.Get("x-auth-token"); !m.DisableAuth && xAuthToken == "" {
			http.Error(w, "Missing x-auth-token header", http.StatusUnauthorized)
			return
		} else if !m.DisableAuth && xAuthToken != m.sessionTokenKey {
			http.Error(w, "Unauthorized: invalid x-auth-token", http.StatusUnauthorized)
			return
		}
		// Extract claims and pass to next handler
		next.ServeHTTP(w, r)
	})
}

func (m *Mock) versionMiddleware(array *Array) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			muxVars := mux.Vars(r)
			apiVersion, ok := muxVars["api_version"]
			if !ok {
				http.Error(w, "Missing api_version in URL", http.StatusBadRequest)
				return
			}
			if !slices.Contains(array.Versions, apiVersion) {
				fmt.Printf("Unsupported API version: %s\n", apiVersion)
				http.Error(w, "Unsupported API version", http.StatusBadRequest)
				return
			}
			// Extract claims and pass to next handler
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Mock) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiToken := r.Header.Get("api-token")

	if apiToken == "" {
		http.Error(w, "Missing api-token header", http.StatusBadRequest)
		return
	} else if apiToken != m.APIToken {
		http.Error(w, "Unauthorized: invalid api-token", http.StatusUnauthorized)
		return
	}

	sessionToken := m.GenerateSessionToken()
	w.Header().Set("x-auth-token", sessionToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body := map[string]interface{}{
		"items": []map[string]string{
			{"username": m.Username},
		},
	}
	json.NewEncoder(w).Encode(body)
}

func GetVersionHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := map[string][]string{
			"version": array.Versions,
		}
		json.NewEncoder(w).Encode(data)
	}
}

func GetVolumesHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumes := []faclient.Volume{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range strings.Split(ids, ",") {
				volume, err := array.GetVolume(id)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
					return
				}
				volumes = append(volumes, *volume)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				volume, err := array.GetVolumeByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
					return
				}
				volumes = append(volumes, volume)
			}
		} else {
			volumes = array.GetVolumes()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		data := faclient.VolumesList{
			Items:              volumes,
			TotalItemCount:     len(volumes),
			MoreItemsRemaining: false,
		}
		json.NewEncoder(w).Encode(data)
	}
}

func PatchVolumeHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumes := []faclient.Volume{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range strings.Split(ids, ",") {
				volume, err := array.GetVolume(id)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
					return
				}
				volumes = append(volumes, *volume)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				volume, err := array.GetVolumeByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
					return
				}
				volumes = append(volumes, volume)
			}
		} else {
			volumes = array.GetVolumes()
		}

		var volumePatch faclient.VolumePatch
		err := json.NewDecoder(r.Body).Decode(&volumePatch)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if volumePatch.Name != nil {
			if len(volumes) > 1 {
				http.Error(w, "Cannot rename multiple volumes at once", http.StatusBadRequest)
				return
			}
		} else {
			for i := range volumes {
				array.UpdateVolume(volumes[i].Id, volumePatch)
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Create or copy a volume and upsert tags
func PostVolumeHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeNames := strings.Split(r.URL.Query().Get("names"), ",")
		if len(volumeNames) < 1 || volumeNames[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		addToPromotionGroupIds := strings.Split(r.URL.Query().Get("add_to_promotion_group_ids"), ",")
		addToPromotionGroupNames := strings.Split(r.URL.Query().Get("add_to_promotion_group_names"), ",")
		if len(addToPromotionGroupNames) > 0 && addToPromotionGroupNames[0] != "" {
			for _, name := range addToPromotionGroupNames {
				pg, err := array.GetProtectionGroupByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Protection group with name %s not found", name), http.StatusNotFound)
					return
				}
				addToPromotionGroupIds = append(addToPromotionGroupIds, pg.Id)
			}
		}
		slices.Sort(addToPromotionGroupIds)
		addToPromotionGroupIds = slices.Compact(addToPromotionGroupIds)

		var volumePost faclient.VolumePost
		err := json.NewDecoder(r.Body).Decode(&volumePost)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		volumeResult := faclient.VolumesList{
			Items: []faclient.Volume{},
		}
		for _, name := range volumeNames {
			newVolume := NewVolumeFromVolumePost(volumePost, name)
			volume, err := array.AddVolume(newVolume)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to add volume: %v", err), http.StatusInternalServerError)
				return
			}
			volumeResult.Items = append(volumeResult.Items, *volume)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(volumeResult)
	}
}

func DeleteVolumeHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumes := []faclient.Volume{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range strings.Split(ids, ",") {
				volume, err := array.GetVolume(id)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume with ID %s not found", id), http.StatusNotFound)
					return
				}
				volumes = append(volumes, *volume)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				volume, err := array.GetVolumeByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume with name %s not found", name), http.StatusNotFound)
					return
				}
				volumes = append(volumes, volume)
			}
		} else {
			http.Error(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, volume := range volumes {
			err := array.EradicateVolume(volume.Id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete volume with ID %s: %v", volume.Id, err), http.StatusBadRequest)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetVolumeGroupsHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vgroups := []faclient.VolumeGroup{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range strings.Split(ids, ",") {
				vgroup, err := array.GetVolumeGroup(id)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume group with ID %s not found", id), http.StatusNotFound)
					return
				}
				vgroups = append(vgroups, *vgroup)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				vgroup, err := array.GetVolumeGroupByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
					return
				}
				vgroups = append(vgroups, *vgroup)
			}
		} else {
			vgroups = array.GetVolumeGroups()
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := faclient.VolumeGroupsList{
			Items:              vgroups,
			TotalItemCount:     len(vgroups),
			MoreItemsRemaining: false,
		}
		json.NewEncoder(w).Encode(data)
	}
}

// Create or copy a volume group and upsert tags
func PostVolumeGroupHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeNames := strings.Split(r.URL.Query().Get("names"), ",")
		if len(volumeNames) < 1 || volumeNames[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var volumeGroupPost faclient.VolumeGroupPost
		err := json.NewDecoder(r.Body).Decode(&volumeGroupPost)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		volumeGroupResult := faclient.VolumeGroupsList{
			Items: []faclient.VolumeGroup{},
		}
		for _, name := range volumeNames {
			newVolumeGroup := NewVolumeGroup(name, &volumeGroupPost)
			vgroup, err := array.AddVolumeGroup(*newVolumeGroup)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to add volume group: %v", err), http.StatusInternalServerError)
				return
			}
			volumeGroupResult.Items = append(volumeGroupResult.Items, *vgroup)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(volumeGroupResult)
	}
}

func PatchVolumeGroupHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeGroups := []faclient.VolumeGroup{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range strings.Split(ids, ",") {
				vgroup, err := array.GetVolumeGroup(id)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume group with ID %s not found", id), http.StatusNotFound)
					return
				}
				volumeGroups = append(volumeGroups, *vgroup)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				vgroup, err := array.GetVolumeGroupByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
					return
				}
				volumeGroups = append(volumeGroups, *vgroup)
			}
		} else {
			http.Error(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		var volumeGroupPatch faclient.VolumeGroupPatch
		err := json.NewDecoder(r.Body).Decode(&volumeGroupPatch)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if volumeGroupPatch.Name != nil {
			if len(volumeGroups) > 1 {
				http.Error(w, "Cannot rename multiple volume groups at once", http.StatusBadRequest)
				return
			}
		}

		vgroupResult := faclient.VolumeGroupsList{
			Items: []faclient.VolumeGroup{},
		}
		for i := range volumeGroups {
			updatedVgroup, err := array.UpdateVolumeGroup(volumeGroups[i].Id, volumeGroupPatch)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update volume group with ID %s: %v", volumeGroups[i].Id, err), http.StatusInternalServerError)
				return
			}
			vgroupResult.Items = append(vgroupResult.Items, *updatedVgroup)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vgroupResult)
	}
}

func DeleteVolumeGroupHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		volumeGroups := []faclient.VolumeGroup{}
		if ids := r.URL.Query().Get("ids"); ids != "" {
			for _, id := range strings.Split(ids, ",") {
				vgroup, err := array.GetVolumeGroup(id)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume group with ID %s not found", id), http.StatusNotFound)
					return
				}
				volumeGroups = append(volumeGroups, *vgroup)
			}
		} else if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				vgroup, err := array.GetVolumeGroupByName(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Volume group with name %s not found", name), http.StatusNotFound)
					return
				}
				volumeGroups = append(volumeGroups, *vgroup)
			}
		} else {
			http.Error(w, "Missing 'ids' query parameter or 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, vgroup := range volumeGroups {
			err := array.EradicateVolumeGroup(vgroup.Id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete volume group with ID %s: %v", vgroup.Id, err), http.StatusBadRequest)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetHostsHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hosts := []faclient.Host{}
		if names := r.URL.Query().Get("names"); names != "" {
			for _, name := range strings.Split(names, ",") {
				host, err := array.GetHost(name)
				if err != nil {
					http.Error(w, fmt.Sprintf("Host with name %s not found", name), http.StatusNotFound)
					return
				}
				hosts = append(hosts, *host)
			}
		} else {
			hosts = array.GetHosts()
		}

		w.Header().Set("Content-Type", "application/json")
		data := faclient.HostsList{
			Items:              hosts,
			TotalItemCount:     len(hosts),
			MoreItemsRemaining: false,
		}
		json.NewEncoder(w).Encode(data)
	}
}

func PostHostHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostNames := strings.Split(r.URL.Query().Get("names"), ",")
		if len(hostNames) < 1 || hostNames[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var hostPost faclient.HostPostBody
		err := json.NewDecoder(r.Body).Decode(&hostPost)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		hostsCreated := []faclient.Host{}
		for _, name := range hostNames {
			newHost := faclient.Host{
				Name:        name,
				Personality: faclient.HostPersonality(hostPost.Personality),
				IQNs:        hostPost.IQNs,
				NQNs:        hostPost.NQNs,
				WWNs:        hostPost.WWNs,
				Vlan:        hostPost.Vlan,
				Chap:        hostPost.Chap,
				Destroyed:   false,
			}

			host, err := array.AddHost(newHost)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to add host: %v", err), http.StatusInternalServerError)
				return
			}
			hostsCreated = append(hostsCreated, *host)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NewResults(hostsCreated))
	}
}

func PatchHostHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := strings.Split(r.URL.Query().Get("names"), ",")
		if len(names) == 0 || names[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var hostPatch faclient.HostPatchBody
		err := json.NewDecoder(r.Body).Decode(&hostPatch)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(names) > 1 && hostPatch.Name != "" {
			http.Error(w, "Cannot specify multiple names in the query parameter when 'name' field is set in the request body", http.StatusBadRequest)
			return
		}

		hosts := []faclient.Host{}
		for _, name := range names {
			host, err := array.UpdateHost(name, hostPatch)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update host: %v", err), http.StatusInternalServerError)
				return
			}
			hosts = append(hosts, *host)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NewResults(hosts))
	}
}

func DeleteHostHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := strings.Split(r.URL.Query().Get("names"), ",")
		if len(names) < 1 || names[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, name := range names {
			err := array.DeleteHost(name)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete host: %v", err), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

type Results[T any] struct {
	Items              []T  `json:"items"`
	TotalItemCount     int  `json:"total_item_count"`
	MoreItemsRemaining bool `json:"more_items_remaining"`
}

func NewResults[T any](items []T) Results[T] {
	return Results[T]{
		Items:              items,
		TotalItemCount:     len(items),
		MoreItemsRemaining: false,
	}
}

func GetHostGroupsHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		hostGroups := array.GetHostGroups()
		data := faclient.HostGroupsList{
			Items:              hostGroups,
			TotalItemCount:     len(hostGroups),
			MoreItemsRemaining: false,
		}
		json.NewEncoder(w).Encode(data)
	}
}

func PostHostGroupHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostGroupNames := strings.Split(r.URL.Query().Get("names"), ",")
		if len(hostGroupNames) < 1 || hostGroupNames[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		hostGroupsCreated := []faclient.HostGroup{}
		for _, name := range hostGroupNames {
			newHostGroup := faclient.HostGroup{
				Name:      name,
				Destroyed: false,
			}

			hostGroup, err := array.AddHostGroup(newHostGroup)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to add host group: %v", err), http.StatusInternalServerError)
				return
			}
			hostGroupsCreated = append(hostGroupsCreated, *hostGroup)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NewResults(hostGroupsCreated))
	}
}

func PatchHostGroupHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := strings.Split(r.URL.Query().Get("names"), ",")
		if len(names) < 1 || names[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		var hostGroupPatch faclient.HostGroupPatchBody
		err := json.NewDecoder(r.Body).Decode(&hostGroupPatch)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if len(names) > 1 && hostGroupPatch.Name != "" {
			http.Error(w, "Cannot specify multiple names in the query parameter when 'name' field is set in the request body", http.StatusBadRequest)
			return
		}

		hostGroups := []faclient.HostGroup{}
		for _, name := range names {
			hostGroup, err := array.UpdateHostGroup(name, hostGroupPatch)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update host group: %v", err), http.StatusInternalServerError)
				return
			}
			hostGroups = append(hostGroups, *hostGroup)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(NewResults(hostGroups))
	}
}

func DeleteHostGroupHandler(array *Array) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names := strings.Split(r.URL.Query().Get("names"), ",")
		if len(names) == 0 || names[0] == "" {
			http.Error(w, "Missing 'names' query parameter", http.StatusBadRequest)
			return
		}

		for _, name := range names {
			err := array.DeleteHostGroup(name)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete host group: %v", err), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
