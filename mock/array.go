package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	faclient "github.com/sanderdescamps/go-purefa"
)

var (
	ErrVolumeNotFound     = fmt.Errorf("volume not found")
	ErrVolumeNotDestroyed = fmt.Errorf("volume is not destroyed, cannot eradicate")
)

type Array struct {
	Versions         faclient.ApiVersions
	Volumes          []*faclient.Volume
	VolumeGroups     []*faclient.VolumeGroup
	Hosts            []*faclient.Host
	HostGroups       []*faclient.HostGroup
	ProtectionGroups []*faclient.ProtectionGroup
	Pods             []*faclient.Pod
}

func NewArray() *Array {
	return &Array{
		Versions:         []string{"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7", "1.8", "1.9", "1.10", "1.11", "1.12", "1.13", "1.14", "1.15", "1.16", "1.17", "1.18", "1.19", "2.0", "2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "2.8", "2.9", "2.10", "2.11", "2.13", "2.14", "2.15", "2.16", "2.17", "2.19", "2.20", "2.21", "2.22", "2.23", "2.24", "2.25", "2.26", "2.27", "2.28", "2.29", "2.30", "2.31", "2.32", "2.33", "2.34", "2.35", "2.36", "2.37", "2.38", "2.39", "2.40", "2.41", "2.42", "2.43", "2.44", "2.45", "2.46"},
		Volumes:          []*faclient.Volume{},
		VolumeGroups:     []*faclient.VolumeGroup{},
		Hosts:            []*faclient.Host{},
		HostGroups:       []*faclient.HostGroup{},
		ProtectionGroups: []*faclient.ProtectionGroup{},
		Pods:             []*faclient.Pod{},
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

	if matcher("volumes")(name) {
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
	} else if matcher("protection_groups")(name) || matcher("protection-groups")(name) {
		items, err := readItemsFromResultBytes[faclient.ProtectionGroup](results)
		if err != nil {
			return err
		}
		array.ProtectionGroups = items
		return nil
	} else if matcher("pods")(name) {
		items, err := readItemsFromResultBytes[faclient.Pod](results)
		if err != nil {
			return err
		}
		array.Pods = items
		return nil
	} else {
		fmt.Fprintf(os.Stderr, "unknown resource type for file %s\n", name)
		return nil
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

// // NewFileLoader loads items from a file into a slice pointer
// // Parameters:
// //   - file: fs.File containing JSON data with a result object
// //   - r: pointer to a slice where results will be loaded
// //
// // Returns error if:
// //   - r is not a pointer
// //   - r does not point to a slice
// //   - file reading fails
// //   - JSON unmarshaling fails
// func NewGenericFileLoader[T any](file fs.File, r *[]*T) error {
// 	results := faclient.Results[T]{
// 		Items: []T{},
// 	}

// 	// Read file content and Unmarshal into result
// 	if err := loadJSONFromFile(file, &results); err != nil {
// 		return err
// 	}

// 	*r = make([]*T, len(results.Items))
// 	for i, item := range results.Items {
// 		(*r)[i] = &item
// 	}
// 	return nil
// }

// func NewGenericInterfaceLoader[T any](results faclient.Results[any], r *[]*T) error {
// 	*r = make([]*T, len(results.Items))
// 	for i, item := range results.Items {
// 		itemBytes, err := json.Marshal(item)
// 		if err != nil {
// 			return fmt.Errorf("marshal item %d: %w", i, err)
// 		}
// 		var typedItem T
// 		if err := json.Unmarshal(itemBytes, &typedItem); err != nil {
// 			return fmt.Errorf("unmarshal item %d: %w", i, err)
// 		}
// 		(*r)[i] = &typedItem
// 	}
// 	return nil
// }

// type Loader func(name string, results faclient.Results[any]) error

// type LoaderFunc func(fileName string, file fs.File) error
type ExporterFunc func(dir string) error

// func ResourceFileLoaderFunc[T any](prefix string, dest *[]*T) Loader {

// }

// func ResourceLoaderFunc[T any](prefix string, dest *[]*T) LoaderFunc {
// 	return func(fileName string, file fs.File) error {
// 		if !strings.HasPrefix(fileName, prefix) {
// 			return nil
// 		}
// 		var results faclient.Results[T]
// 		if err := loadJSONFromFile(file, &results); err != nil {
// 			return err
// 		}
// 		items := make([]*T, len(results.Items))
// 		for i, item := range results.Items {
// 			items[i] = &item
// 		}

// 		*dest = items
// 		return nil
// 	}
// }

func ResourceExporterFunc[T any](fileName string, resource *[]*T) ExporterFunc {
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

// func (array *Array) Loader() LoaderFunc {
// 	ResourceLoaders := []func(fileName string, file fs.File) error{}
// 	ResourceLoaders = append(ResourceLoaders, ResourceLoaderFunc("volumes", &array.Volumes))
// 	ResourceLoaders = append(ResourceLoaders, ResourceLoaderFunc("volume_groups", &array.VolumeGroups))
// 	ResourceLoaders = append(ResourceLoaders, ResourceLoaderFunc("hosts", &array.Hosts))
// 	ResourceLoaders = append(ResourceLoaders, ResourceLoaderFunc("host_groups", &array.HostGroups))
// 	ResourceLoaders = append(ResourceLoaders, ResourceLoaderFunc("protection_groups", &array.ProtectionGroups))
// 	ResourceLoaders = append(ResourceLoaders, ResourceLoaderFunc("pods", &array.Pods))

// 	return func(fileName string, file fs.File) error {
// 		for _, loader := range ResourceLoaders {
// 			if err := loader(fileName, file); err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	}
// }

// func (array *Array) Export(dir string) error {
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return fmt.Errorf("create output directory %s: %w", dir, err)
// 	}

// 	exporters := []ExporterFunc{}
// 	exporters = append(exporters, ResourceExporterFunc("volumes.json", &array.Volumes))
// 	exporters = append(exporters, ResourceExporterFunc("volume_groups.json", &array.VolumeGroups))
// 	exporters = append(exporters, ResourceExporterFunc("hosts.json", &array.Hosts))
// 	exporters = append(exporters, ResourceExporterFunc("host_groups.json", &array.HostGroups))
// 	exporters = append(exporters, ResourceExporterFunc("protection_groups.json", &array.ProtectionGroups))
// 	exporters = append(exporters, ResourceExporterFunc("pods.json", &array.Pods))

// 	for _, exporter := range exporters {
// 		if err := exporter(dir); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func NewArrayFromFS(fsys fs.FS, root string) (*Array, error) {
	array := NewArray()
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		if err := array.LoadFromBytes(filepath.Base(path), data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load mock data from FS %s: %w", root, err)
	}
	return array, nil
}

func NewArrayFromDir(dir string) (*Array, error) {

	array := NewArray()

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

		if err := array.LoadFromBytes(filepath.Base(path), data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load mock data from directory %s: %w", dir, err)
	}
	return array, nil
}

func NewArrayFromClient(client *faclient.FAClient) (*Array, error) {
	volumes, err := client.GetVolumes()
	if err != nil {
		return nil, err
	}
	volumeGroups, err := client.GetVolumeGroups()
	if err != nil {
		return nil, err
	}
	hosts, err := client.GetHosts()
	if err != nil {
		return nil, err
	}
	hostGroups, err := client.GetHostGroups()
	if err != nil {
		return nil, err
	}
	// protectionGroups, err := client.GetProtectionGroups()
	// if err != nil {
	// 	return nil, err
	// }
	// pods, err := client.GetPods()
	// if err != nil {
	// 	return nil, err
	// }

	versions, err := client.GetVersions()
	if err != nil {
		return nil, err
	}

	return &Array{
		Versions:     versions,
		Volumes:      SliceToPointerSlice(volumes),
		VolumeGroups: SliceToPointerSlice(volumeGroups),
		Hosts:        SliceToPointerSlice(hosts),
		HostGroups:   SliceToPointerSlice(hostGroups),
		// ProtectionGroups: SliceToPointerSlice(protectionGroups),
		// Pods: SliceToPointerSlice(pods),
	}, nil

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

func (array *Array) ExportToDir(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("create output directory %s: %w", dirPath, err)
	}

	exporters := []ExporterFunc{}
	exporters = append(exporters, ResourceExporterFunc("volumes.json", &array.Volumes))
	exporters = append(exporters, ResourceExporterFunc("volume_groups.json", &array.VolumeGroups))
	exporters = append(exporters, ResourceExporterFunc("hosts.json", &array.Hosts))
	exporters = append(exporters, ResourceExporterFunc("host_groups.json", &array.HostGroups))
	exporters = append(exporters, ResourceExporterFunc("protection_groups.json", &array.ProtectionGroups))
	exporters = append(exporters, ResourceExporterFunc("pods.json", &array.Pods))

	for _, exporter := range exporters {
		if err := exporter(dirPath); err != nil {
			return err
		}
	}

	return nil
}
