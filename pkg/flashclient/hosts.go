package flashclient

import (
	"fmt"
	"net/http"
)

type Chap struct {
	HostPassword   string `json:"host_password"`
	HostUser       string `json:"host_user"`
	TargetPassword string `json:"target_password"`
	TargetUser     string `json:"target_user"`
}

type PortConnectivity struct {
	Details string `json:"details"`
	Status  string `json:"status"`
}

type HostShort struct {
	NoIdReference
}

type Host struct {
	HostShort
	CHAP             Chap             `json:"chap"`
	ConnectionCount  int              `json:"connection_count"`
	Destroyed        bool             `json:"destroyed,omitempty"`
	HostGroup        HostGroupShort   `json:"host_group"`
	IQNs             []string         `json:"iqns"`
	NQNs             []string         `json:"nqns"`
	Personality      HostPersonality  `json:"personality"`
	PortConnectivity PortConnectivity `json:"port_connectivity"`
	PreferredArrays  []ArrayShort     `json:"preferred_arrays"`
	Space            Space            `json:"space"`
	TimeRemaining    int64            `json:"time_remaining,omitempty"`
	VLAN             string           `json:"vlan,omitempty"`
	WWNs             []string         `json:"wwns"`
	IsLocal          bool             `json:"is_local"`
}

type HostPersonality string

const (
	None           HostPersonality = ""
	AIX            HostPersonality = "aix"
	ESXi           HostPersonality = "esxi"
	HitachiVSP     HostPersonality = "hitachi-vsp"
	HPUX           HostPersonality = "hpux"
	OracleVMServer HostPersonality = "oracle-vm-server"
	Solaris        HostPersonality = "solaris"
	VMS            HostPersonality = "vms"
)

type HostPostBody struct {
	CHAP        *Chap            `json:"chap,omitempty"`
	IQNs        []string         `json:"iqns,omitempty"`
	NQNs        []string         `json:"nqns,omitempty"`
	Personality *HostPersonality `json:"personality,omitempty"`
	VLAN        *string          `json:"vlan,omitempty"`
	WWNs        []string         `json:"wwns,omitempty"`
}

type HostPatchBody struct {
	Name *string `json:"name,omitempty"`
	// AddIQNs adds iSCSI Qualified Names to the host
	AddIQNs []string `json:"add_iqns,omitempty"`

	// AddNQNs adds NVMe Qualified Names to the host
	AddNQNs []string `json:"add_nqns,omitempty"`

	// AddWWNs adds Fibre Channel World Wide Names to the host
	AddWWNs []string `json:"add_wwns,omitempty"`

	// CHAP configuration for iSCSI authentication
	CHAP *Chap `json:"chap,omitempty"`

	// HostGroup associates the host with a host group
	HostGroup *HostGroupShort `json:"host_group,omitempty"`

	// IQNs are iSCSI Qualified Names associated with the host
	IQNs []string `json:"iqns,omitempty"`

	// NQNs are NVMe Qualified Names associated with the host
	NQNs []string `json:"nqns,omitempty"`

	// Personality tunes the array for specific host OS/hypervisor
	// Valid values: aix, esxi, hitachi-vsp, hpux, oracle-vm-server, solaris, vms
	Personality *HostPersonality `json:"personality,omitempty"`

	// RemoveIQNs disassociates iSCSI Qualified Names from the host
	RemoveIQNs []string `json:"remove_iqns,omitempty"`

	// RemoveNQNs disassociates NVMe Qualified Names from the host
	RemoveNQNs []string `json:"remove_nqns,omitempty"`

	// RemoveWWNs disassociates Fibre Channel World Wide Names from the host
	RemoveWWNs []string `json:"remove_wwns,omitempty"`

	// VLAN ID for host access (1-4094, "any", or "untagged")
	VLAN *string `json:"vlan,omitempty"`

	// WWNs are Fibre Channel World Wide Names associated with the host
	WWNs []string `json:"wwns,omitempty"`
}

func (fa *FAClient) GetHosts() ([]Host, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/hosts"
	result := Results[Host]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get hosts, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) GetHost(name string) (*Host, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/hosts"
	result := Results[Host]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		SetQueryParam("names", name).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get host %s, got status code %d", name, res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("host %s not found", name)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) CreateHost(name string, body HostPostBody) (*Host, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	result := Results[Host]{}
	res, err := fa.RestClient.R().
		SetBody(body).
		SetResult(&result).
		SetQueryParam("names", name).
		Post("/hosts")
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create host %s, got status code %d", name, res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("host %s not found after creation", name)
	}

	return &result.Items[0], nil
}

func (fa *FAClient) DeleteHost(name string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	res, err := fa.RestClient.R().
		SetQueryParam("names", name).
		Delete("/hosts")
	if err != nil {
		return err
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete host %s, got status code %d", name, res.StatusCode())
	}

	return nil
}

func (fa *FAClient) UpdateHost(name string, body HostPatchBody) (*Host, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	result := Results[Host]{}
	res, err := fa.RestClient.R().
		SetBody(body).
		SetResult(&result).
		SetQueryParam("names", name).
		Patch("/hosts")
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to update host %s, got status code %d", name, res.StatusCode())
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("host %s not found after update", name)
	}

	return &result.Items[0], nil
}
