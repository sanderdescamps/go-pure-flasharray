package fakearray

import faclient "github.com/sanderdescamps/go-purefa"

func (array *Array) GetPorts(filters ...func(*faclient.Port) bool) []faclient.Port {
	ports := []faclient.Port{}
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
