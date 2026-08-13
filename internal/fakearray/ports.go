package fakearray

import "github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"

func (array *Array) GetPorts(filters ...func(*flashclient.Port) bool) []flashclient.Port {
	ports := []flashclient.Port{}
	for _, port := range array.Ports {
		matches := true
		for _, filter := range filters {
			if !filter(port) {
				matches = false
				break
			}
		}
		if matches {
			ports = append(ports, *port)
		}
	}
	return ports
}
