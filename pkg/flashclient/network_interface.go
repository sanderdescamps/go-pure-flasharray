package flashclient

import (
	"fmt"
	"net/http"
)

type NetworkInterface struct {
	NetworkInterfaceShort

	// Returns a value of `true` if the specified network interface or Fibre Channel port is enabled. Returns a value of `false` if the specified network interface or Fibre Channel port is disabled.
	Enabled bool `json:"enabled,omitempty"`

	// Ethernet network interface properties.
	Eth Ethernet `json:"eth,omitempty"`

	// Fibre Channel port properties.
	Fc FibreChannel `json:"fc,omitempty"`

	// The interface type. Valid values are `eth` and `fc`.
	InterfaceType string `json:"interface_type,omitempty"`

	// The services provided by the specified network interface or Fibre Channel port.
	Services []string `json:"services,omitempty"`

	// Configured speed of the specified network interface or Fibre Channel port (in Gb/s). Typically this is the maximum speed of the port or bond represented by the network interface.
	Speed int64 `json:"speed,omitempty"`
}

type NetworkInterfaceShort struct {
	// A locally unique, system-generated name. The name cannot be modified.
	Name string `json:"name,omitempty"`
}

type Subnet struct {
	Name string `json:"name"`
}

type EthernetSubtype struct {
	Name string `json:"name"`
}

// Ethernet - Ethernet network interface properties.
type Ethernet struct {

	// The IPv4 or IPv6 address to be associated with the specified network interface.
	Address string `json:"address,omitempty"`

	// The IPv4 or IPv6 address of the gateway through which the specified network interface is to communicate with the network.
	Gateway string `json:"gateway,omitempty"`

	// The media access control address associated with the specified network interface.
	MacAddress string `json:"mac_address,omitempty"`

	// Maximum message transfer unit (packet) size for the network interface, in bytes. MTU setting cannot exceed the MTU of the corresponding physical interface.
	Mtu int32 `json:"mtu,omitempty"`

	// Netmask of the specified network interface that, when combined with the address of the interface, determines the network address of the interface.
	Netmask string `json:"netmask,omitempty"`

	// List of network interfaces configured to be a subinterface of the specified network interface.
	Subinterfaces []NetworkInterfaceShort `json:"subinterfaces,omitempty"`

	// Subnet that is associated with the specified network interface.
	Subnet Subnet `json:"subnet,omitempty"`

	// The subtype of the specified network interface. Only interfaces of subtype `virtual` can be created. Configurable on POST only. Valid values are `failover_bond`, `lacp_bond`, `physical`, and `virtual`.
	Subtype string `json:"subtype,omitempty"`

	// VLAN ID
	Vlan int32 `json:"vlan,omitempty"`
}

// FibreChannel - Fibre Channel port properties.
type FibreChannel struct {

	// World Wide Name of the specified Fibre Channel port.
	Wwn string `json:"wwn,omitempty"`
}

func (fa *FAClient) GetNetworkInterfaces() ([]NetworkInterface, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/network-interfaces"
	result := Results[NetworkInterface]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get network interfaces, got status code %d", res.StatusCode())
	}
	return result.Items, nil
}
