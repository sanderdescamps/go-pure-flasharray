package fakearray

import (
	"slices"

	"github.com/sanderdescamps/go-purefa-mock/pkg/flashclient"
)

func ConnectionsWithHostGroupNames(hostGroupNames ...string) func(*flashclient.Connection) bool {
	return func(c *flashclient.Connection) bool {
		return slices.Contains(hostGroupNames, c.HostGroup.Name)
	}
}

func ConnectionsWithHostNames(hostNames ...string) func(*flashclient.Connection) bool {
	return func(c *flashclient.Connection) bool {
		return slices.Contains(hostNames, c.Host.Name)
	}
}

func ConnectionsWithVolumeIds(volumeIds ...string) func(*flashclient.Connection) bool {
	return func(c *flashclient.Connection) bool {
		return slices.Contains(volumeIds, c.Volume.Id)
	}
}

func (array *Array) nextAvailableLun(hostName string) *int64 {
	for lun := int64(255); lun >= 1; lun-- {
		available := true
		for _, connection := range array.Connections {
			if connection.Host.Name == hostName && connection.Lun != nil && *connection.Lun == lun {
				available = false
				break
			}
		}
		if available {
			return &lun
		}
	}

	for lun := int64(255); lun <= 4095; lun++ {
		available := true
		for _, connection := range array.Connections {
			if connection.Host.Name == hostName && connection.Lun != nil && *connection.Lun == lun {
				available = false
				break
			}
		}
		if available {
			return &lun
		}
	}
	return nil
}

func (array *Array) GetConnections() []flashclient.Connection {
	connections := make([]flashclient.Connection, len(array.Connections))
	for i, connection := range array.Connections {
		connections[i] = *connection
	}

	return connections
}

func (array *Array) AddConnection(connection flashclient.Connection) (*flashclient.Connection, error) {
	if array.ConnectionExists(connection.Host.Name, connection.Volume.Id) {
		return nil, ErrAlreadyExists
	}

	array.Connections = append(array.Connections, &connection)
	return &connection, nil
}

func (array *Array) ConnectionExists(hostName string, volumeId string) bool {
	for _, connection := range array.Connections {
		if connection.Host.Name == hostName && connection.Volume.Id == volumeId {
			return true
		}
	}
	return false
}

func (array *Array) GetConnectionsWithFilter(filter ...func(*flashclient.Connection) bool) []*flashclient.Connection {
	var filteredConnections []*flashclient.Connection
	for _, connection := range array.Connections {
		matches := true
		for _, f := range filter {
			if !f(connection) {
				matches = false
				break
			}
		}
		if matches {
			filteredConnections = append(filteredConnections, connection)
		}
	}
	return filteredConnections
}

func (array *Array) CreateConnection(hostName string, volumeId string, post flashclient.ConnectionPostBody) (*flashclient.Connection, error) {
	host, err := array.GetHost(hostName)
	if err != nil {
		return nil, err
	}

	volume, err := array.GetVolume(volumeId)
	if err != nil {
		return nil, err
	}

	var lun *int64 = nil
	if post.Lun != nil {
		lun = post.Lun
	} else if len(host.IQNs) > 0 && host.IQNs[0] != "" {
		lun = array.nextAvailableLun(hostName)
	}

	connection := flashclient.Connection{
		Host:      host.HostShort,
		Volume:    volume.VolumeShort,
		HostGroup: host.HostGroup,
		Lun: func() *int64 {
			if lun == nil {
				return nil
			}
			v := int64(*lun)
			return &v
		}(),
	}

	if post.ProtocolEndpoint != nil {
		connection.ProtocolEndpoint = *post.ProtocolEndpoint
	}

	return array.AddConnection(connection)
}

func (array *Array) DeleteConnection(hostName string, volumeId string) error {
	for i, connection := range array.Connections {
		if connection.Host.Name == hostName && connection.Volume.Id == volumeId {
			array.Connections = append(array.Connections[:i], array.Connections[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
