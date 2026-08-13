package fakearray

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	faclient "github.com/sanderdescamps/go-purefa"
	"github.com/sanderdescamps/go-purefa-mock/pkg/testdata"
)

var (
	ErrVolumeNotFound      = fmt.Errorf("volume not found")
	ErrVolumeNotDestroyed  = fmt.Errorf("volume is not destroyed, cannot eradicate")
	ErrUnknownResourceType = fmt.Errorf("unknown resource type")
)

type Array struct {
	Versions faclient.ApiVersions

	Alerts                           []*faclient.Alert
	Arrays                           []*faclient.Array
	Controllers                      []*faclient.Controller
	Connections                      []*faclient.Connection
	Drives                           []*faclient.Drive
	Hardware                         []*faclient.Hardware
	Hosts                            []*faclient.Host
	HostGroups                       []*faclient.HostGroup
	NetworkInterfaces                []*faclient.NetworkInterface
	Pods                             []*faclient.Pod
	Ports                            []*faclient.Port
	ProtectionGroups                 []*faclient.ProtectionGroup
	ProtectionGroupsHostMembers      []*faclient.ProtectionGroupHostMember
	ProtectionGroupsHostGroupMembers []*faclient.ProtectionGroupHostGroupMember
	ProtectionGroupsVolumeMembers    []*faclient.ProtectionGroupVolumeMember
	ProtectionGroupSnapshots         []*faclient.ProtectionGroupSnapshot
	Volumes                          []*faclient.Volume
	VolumeGroups                     []*faclient.VolumeGroup
	VolumeSnapshots                  []*faclient.VolumeSnapshot

	logger *slog.Logger
}

type ArrayOption func(*Array)

func WithLogger(logger *slog.Logger) ArrayOption {
	return func(array *Array) {
		array.logger = logger
	}
}

func NewArray(arrayOptions ...ArrayOption) *Array {
	array := &Array{
		Versions:                         []string{"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9", "1.10", "1.11", "1.12", "1.13", "1.14", "1.15", "1.16", "1.17", "1.18", "1.19", "2.0", "2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "2.8", "2.9", "2.10", "2.11", "2.13", "2.14", "2.15", "2.16", "2.17", "2.19", "2.20", "2.21", "2.22", "2.23", "2.24", "2.25", "2.26", "2.27", "2.28", "2.29", "2.30", "2.31", "2.32", "2.33", "2.34", "2.35", "2.36", "2.37", "2.38", "2.39", "2.40", "2.41", "2.42", "2.43", "2.44", "2.45", "2.46"},
		Alerts:                           []*faclient.Alert{},
		Arrays:                           []*faclient.Array{},
		Controllers:                      []*faclient.Controller{},
		Connections:                      []*faclient.Connection{},
		Drives:                           []*faclient.Drive{},
		Hardware:                         []*faclient.Hardware{},
		Hosts:                            []*faclient.Host{},
		HostGroups:                       []*faclient.HostGroup{},
		NetworkInterfaces:                []*faclient.NetworkInterface{},
		ProtectionGroups:                 []*faclient.ProtectionGroup{},
		ProtectionGroupsHostMembers:      []*faclient.ProtectionGroupHostMember{},
		ProtectionGroupsHostGroupMembers: []*faclient.ProtectionGroupHostGroupMember{},
		ProtectionGroupsVolumeMembers:    []*faclient.ProtectionGroupVolumeMember{},
		Pods:                             []*faclient.Pod{},
		Volumes:                          []*faclient.Volume{},
		VolumeGroups:                     []*faclient.VolumeGroup{},
		VolumeSnapshots:                  []*faclient.VolumeSnapshot{},
		logger:                           slog.Default(),
	}
	for _, option := range arrayOptions {
		option(array)
	}
	return array
}

func (array *Array) SetLogger(logger *slog.Logger) {
	array.logger = logger
}

func (array *Array) logDebug(msg string, a ...any) {
	if array.logger != nil {
		array.logger.Debug(fmt.Sprintf(msg, a...))
	}
}

func (array *Array) logInfo(msg string, a ...any) {
	if array.logger != nil {
		array.logger.Info(fmt.Sprintf(msg, a...))
	}
}

func (array *Array) logWarn(msg string, a ...any) {
	if array.logger != nil {
		array.logger.Warn(fmt.Sprintf(msg, a...))
	}
}

func (array *Array) logError(msg string, a ...any) {
	if array.logger != nil {
		array.logger.Error(fmt.Sprintf(msg, a...))
	}
}

func (array *Array) LoadFromBytes(name string, results []byte) error {
	matcher := func(name string) func(filename string) bool {
		return func(filename string) bool {
			filename = strings.Split(filename, ".")[0]
			if strings.EqualFold(filename, name) {
				return true
			}
			return false
		}
	}

	if matcher("versions")(name) {
		body := faclient.VersionsResponse{}
		err := json.Unmarshal(results, &body)
		if err != nil {
			return err
		}
		array.Versions = body.Versions
		return nil
	}

	if matcher("alerts")(name) {
		items, err := readItemsFromResultBytes[faclient.Alert](results)
		if err != nil {
			return err
		}
		array.Alerts = items
		return nil
	} else if matcher("arrays")(name) {
		items, err := readItemsFromResultBytes[faclient.Array](results)
		if err != nil {
			return err
		}
		array.Arrays = items
		return nil
	} else if matcher("connections")(name) {
		items, err := readItemsFromResultBytes[faclient.Connection](results)
		if err != nil {
			return err
		}
		array.Connections = items
		return nil
	} else if matcher("controllers")(name) {
		items, err := readItemsFromResultBytes[faclient.Controller](results)
		if err != nil {
			return err
		}
		array.Controllers = items
		return nil
	} else if matcher("drives")(name) {
		items, err := readItemsFromResultBytes[faclient.Drive](results)
		if err != nil {
			return err
		}
		array.Drives = items
		return nil
	} else if matcher("hardware")(name) {
		items, err := readItemsFromResultBytes[faclient.Hardware](results)
		if err != nil {
			return err
		}
		array.Hardware = items
		return nil
	} else if matcher("hosts")(name) {
		items, err := readItemsFromResultBytes[faclient.Host](results)
		if err != nil {
			return err
		}
		array.Hosts = items
		return nil
	} else if matcher("host_groups")(name) || matcher("host-groups")(name) {
		items, err := readItemsFromResultBytes[faclient.HostGroup](results)
		if err != nil {
			return err
		}
		array.HostGroups = items
		return nil
	} else if matcher("network_interfaces")(name) || matcher("network-interfaces")(name) {
		items, err := readItemsFromResultBytes[faclient.NetworkInterface](results)
		if err != nil {
			return err
		}
		array.NetworkInterfaces = items
		return nil
	} else if matcher("pods")(name) {
		items, err := readItemsFromResultBytes[faclient.Pod](results)
		if err != nil {
			return err
		}
		array.Pods = items
		return nil
	} else if matcher("ports")(name) {
		items, err := readItemsFromResultBytes[faclient.Port](results)
		if err != nil {
			return err
		}
		array.Ports = items
		return nil
	} else if matcher("protection_groups")(name) || matcher("protection-groups")(name) {
		items, err := readItemsFromResultBytes[faclient.ProtectionGroup](results)
		if err != nil {
			return err
		}
		array.ProtectionGroups = items
		return nil
	} else if matcher("protection_groups_members_host")(name) || matcher("protection-groups-members-host")(name) || matcher("protection_groups_hosts")(name) || matcher("protection-groups-hosts")(name) {
		items, err := readItemsFromResultBytes[faclient.ProtectionGroupHostMember](results)
		if err != nil {
			return err
		}
		array.ProtectionGroupsHostMembers = items
		return nil
	} else if matcher("protection_groups_members_host_group")(name) || matcher("protection-groups-members-host-group")(name) || matcher("protection_groups_host_groups")(name) || matcher("protection-groups-host-groups")(name) {
		items, err := readItemsFromResultBytes[faclient.ProtectionGroupHostGroupMember](results)
		if err != nil {
			return err
		}
		array.ProtectionGroupsHostGroupMembers = items
		return nil
	} else if matcher("protection_groups_members_volume")(name) || matcher("protection-groups-members-volume")(name) || matcher("protection_groups_volumes")(name) || matcher("protection-groups-volumes")(name) {
		items, err := readItemsFromResultBytes[faclient.ProtectionGroupVolumeMember](results)
		if err != nil {
			return err
		}
		array.ProtectionGroupsVolumeMembers = items
		return nil
	} else if matcher("volumes")(name) {
		items, err := readItemsFromResultBytes[faclient.Volume](results)
		if err != nil {
			return err
		}
		for _, volume := range items {
			_, err := array.AddVolume(*volume)
			if errors.Is(err, ErrAlreadyExists) {
				vol, _ := array.GetVolume(volume.Id)
				*vol = *volume
			} else if err != nil {
				return err
			}
		}
		return nil
	} else if matcher("volume_groups")(name) || matcher("volume-groups")(name) {
		items, err := readItemsFromResultBytes[faclient.VolumeGroup](results)
		if err != nil {
			return err
		}
		for _, item := range items {
			_, err := array.AddVolumeGroup(*item)
			if errors.Is(err, ErrAlreadyExists) {
				vg, _ := array.GetVolumeGroup(item.Id)
				*vg = *item
			} else if err != nil {
				return err
			}
		}
		return nil
	} else {
		return fmt.Errorf("unknown resource type for file %s: %w", name, ErrUnknownResourceType)
	}
}

func readItemsFromResultBytes[T any](results []byte) ([]*T, error) {
	var res faclient.Results[T]
	err := json.Unmarshal(results, &res)
	if err != nil {
		return nil, err
	}

	items := make([]*T, len(res.Items))
	for i, item := range res.Items {
		items[i] = &item
	}
	return items, nil
}

type SaveFunc func(dir string) error

func ResourceSaveFunc[T any](fileName string, resource *[]*T) SaveFunc {
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}
	return func(dir string) error {
		results := faclient.NewResults(*resource)
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", fileName, err)
		}
		path := filepath.Join(dir, fileName)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write result for %s: %w", fileName, err)
		}
		return nil
	}
}

func VersionsSaveFunc(fileName string, versions faclient.ApiVersions) SaveFunc {
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}
	return func(dir string) error {
		body := faclient.VersionsResponse{
			Versions: versions,
		}
		data, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", fileName, err)
		}
		path := filepath.Join(dir, fileName)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write result for %s: %w", fileName, err)
		}
		return nil
	}
}

func NewArrayWithTestData(arrayOptions ...ArrayOption) (*Array, error) {
	testData := testdata.TestData
	array := NewArray()
	err := fs.WalkDir(testData, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		file, err := testData.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		err = array.LoadFromBytes(filepath.Base(path), data)
		if err != nil && !errors.Is(err, ErrUnknownResourceType) {
			return fmt.Errorf("failed to load mock data from file %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load mock data from FS: %w", err)
	}
	return array, nil
}

func NewArrayFromDir(dir string, arrayOptions ...ArrayOption) (*Array, error) {

	array := NewArray(arrayOptions...)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		err = array.LoadFromBytes(filepath.Base(path), data)
		if errors.Is(err, ErrUnknownResourceType) {
			array.logWarn("skipping unknown resource type for file %s", path)
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to load mock data from file %s: %w", path, err)
		}
		array.logDebug("successfully loaded data from %s", path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load mock data from directory %s: %w", dir, err)
	}
	return array, nil
}

func NewArrayFromClient(client *faclient.FAClient, arrayOptions ...ArrayOption) (*Array, error) {

	hosts, err := client.GetHosts()
	if err != nil {
		return nil, err
	}
	hostGroups, err := client.GetHostGroups()
	if err != nil {
		return nil, err
	}
	protectionGroups, err := client.GetProtectionGroups()
	if err != nil {
		return nil, err
	}
	protectionGroupsHostMembers, err := client.GetAllProtectionGroupHostMembers()
	if err != nil {
		return nil, err
	}
	protectionGroupsHostGroupMembers, err := client.GetAllProtectionGroupHostGroupMembers()
	if err != nil {
		return nil, err
	}
	protectionGroupsVolumeMembers, err := client.GetAllProtectionGroupVolumeMembers()
	if err != nil {
		return nil, err
	}

	pods, err := client.GetPods()
	if err != nil {
		return nil, err
	}
	volumes, err := client.GetVolumes()
	if err != nil {
		return nil, err
	}
	volumeGroups, err := client.GetVolumeGroups()
	if err != nil {
		return nil, err
	}
	volumeSnapshots, err := client.GetVolumeSnapshots()
	if err != nil {
		return nil, err
	}

	versions, err := client.GetVersions()
	if err != nil {
		return nil, err
	}

	array := &Array{
		Versions:                         versions,
		Volumes:                          sliceToPointerSlice(volumes),
		VolumeGroups:                     sliceToPointerSlice(volumeGroups),
		Hosts:                            sliceToPointerSlice(hosts),
		HostGroups:                       sliceToPointerSlice(hostGroups),
		ProtectionGroups:                 sliceToPointerSlice(protectionGroups),
		ProtectionGroupsHostMembers:      sliceToPointerSlice(protectionGroupsHostMembers),
		ProtectionGroupsHostGroupMembers: sliceToPointerSlice(protectionGroupsHostGroupMembers),
		ProtectionGroupsVolumeMembers:    sliceToPointerSlice(protectionGroupsVolumeMembers),
		Pods:                             sliceToPointerSlice(pods),
		VolumeSnapshots:                  sliceToPointerSlice(volumeSnapshots),
	}
	for _, option := range arrayOptions {
		option(array)
	}
	return array, nil

}

func CorrectMissingReferences(array *Array) error {
	// Validate Volumes
	for _, volume := range array.Volumes {
		if volume.VolumeGroup != nil {
			if volume.VolumeGroup.Id == "" {
				volume.VolumeGroup = nil
			} else if _, err := array.GetVolumeGroup(volume.VolumeGroup.Id); err != nil {
				newVGroup := NewVolumeGroup(volume.VolumeGroup.Name, nil)
				array.AddVolumeGroup(*newVGroup)
			}
		}
	}

	for _, host := range array.Hosts {
		if host.HostGroup.Name != "" {
			if _, err := array.GetHostGroup(host.HostGroup.Name); err != nil {
				newHGroup := NewHostGroup(host.HostGroup.Name)
				array.AddHostGroup(*newHGroup)
			}
		}
	}

	return nil
}

func (array *Array) SaveToDir(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("create output directory %s: %w", dirPath, err)
	}

	exporters := []SaveFunc{}
	exporters = append(exporters, VersionsSaveFunc("versions.json", array.Versions))
	exporters = append(exporters, ResourceSaveFunc("alerts.json", &array.Alerts))
	exporters = append(exporters, ResourceSaveFunc("arrays.json", &array.Arrays))
	exporters = append(exporters, ResourceSaveFunc("controllers.json", &array.Controllers))
	exporters = append(exporters, ResourceSaveFunc("connections.json", &array.Connections))
	exporters = append(exporters, ResourceSaveFunc("hosts.json", &array.Hosts))
	exporters = append(exporters, ResourceSaveFunc("host_groups.json", &array.HostGroups))
	exporters = append(exporters, ResourceSaveFunc("network_interfaces.json", &array.NetworkInterfaces))
	exporters = append(exporters, ResourceSaveFunc("protection_groups.json", &array.ProtectionGroups))
	exporters = append(exporters, ResourceSaveFunc("protection_groups_members_host.json", &array.ProtectionGroupsHostMembers))
	exporters = append(exporters, ResourceSaveFunc("protection_groups_members_host_group.json", &array.ProtectionGroupsHostGroupMembers))
	exporters = append(exporters, ResourceSaveFunc("protection_groups_members_volume.json", &array.ProtectionGroupsVolumeMembers))
	exporters = append(exporters, ResourceSaveFunc("pods.json", &array.Pods))
	exporters = append(exporters, ResourceSaveFunc("volumes.json", &array.Volumes))
	exporters = append(exporters, ResourceSaveFunc("volume_groups.json", &array.VolumeGroups))
	exporters = append(exporters, ResourceSaveFunc("volume_snapshots.json", &array.VolumeSnapshots))

	for _, exporter := range exporters {
		if err := exporter(dirPath); err != nil {
			return err
		}
	}

	return nil
}

func (array *Array) GetArrays() ([]faclient.Array, error) {
	arrays := make([]faclient.Array, len(array.Arrays))
	for i, arr := range array.Arrays {
		arrays[i] = *arr
	}
	return arrays, nil
}

func (array *Array) GetVersions() faclient.ApiVersions {
	return array.Versions
}
