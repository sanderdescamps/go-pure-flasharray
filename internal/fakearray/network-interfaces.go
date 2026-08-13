package fakearray

import (
	"fmt"

	faclient "github.com/sanderdescamps/go-purefa"
)

func (array *Array) GetNetworkInterfaces(filters ...func(*faclient.NetworkInterface) bool) []faclient.NetworkInterface {
	networkInterfaces := []faclient.NetworkInterface{}
	for _, ni := range array.NetworkInterfaces {
		matches := true
		for _, filter := range filters {
			if !filter(ni) {
				matches = false
				break
			}
		}
		if matches {
			networkInterfaces = append(networkInterfaces, *ni)
		}
	}
	return networkInterfaces
}

func (array *Array) GetNetworkInterface(name string) (*faclient.NetworkInterface, error) {
	for _, ni := range array.NetworkInterfaces {
		if ni.Name == name {
			return ni, nil
		}
	}
	return nil, fmt.Errorf("network interface with name %s not found", name)
}
