package fakearray

import faclient "github.com/sanderdescamps/go-purefa"

func (array *Array) GetDrives(filters ...func(*faclient.Drive) bool) []faclient.Drive {
	drives := []faclient.Drive{}
	for _, drive := range array.Drives {
		matches := true
		for _, filter := range filters {
			if !filter(drive) {
				matches = false
				break
			}
		}
		if matches {
			drives = append(drives, *drive)
		}
	}
	return drives
}
