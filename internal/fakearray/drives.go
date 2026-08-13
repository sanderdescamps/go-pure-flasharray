package fakearray

import "github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"

func (array *Array) GetDrives(filters ...func(*flashclient.Drive) bool) []flashclient.Drive {
	drives := []flashclient.Drive{}
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
