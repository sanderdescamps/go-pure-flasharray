package mock

import (
	"fmt"
	"slices"

	faclient "github.com/sanderdescamps/go-purefa"
)

func (array *Array) GetHosts() []faclient.Host {
	hosts := make([]faclient.Host, len(array.Hosts))
	for i, host := range array.Hosts {
		hosts[i] = *host
	}

	return hosts
}

func (array *Array) GetHost(name string) (*faclient.Host, error) {
	for _, host := range array.Hosts {
		if host.Name == name {
			return host, nil
		}
	}
	return nil, fmt.Errorf("host with name %s not found", name)
}

func (array *Array) AddHost(host faclient.Host) (*faclient.Host, error) {
	for _, h := range array.Hosts {
		if h.Name == host.Name {
			return nil, fmt.Errorf("host with name %s already exists", host.Name)
		}
	}
	array.Hosts = append(array.Hosts, &host)
	return &host, nil
}

func (array *Array) UpdateHost(name string, hostPatch faclient.HostPatchBody) (*faclient.Host, error) {
	for i, host := range array.Hosts {
		if host.Name == name {
			if hostPatch.Personality != nil {
				host.Personality = *hostPatch.Personality
			}
			if hostPatch.VLAN != nil {
				host.Vlan = *hostPatch.VLAN
			}

			//IQN
			if len(hostPatch.IQNs) > 0 {
				host.IQNs = hostPatch.IQNs
			}
			if len(hostPatch.RemoveIQNs) > 0 {
				// Remove IQNs from the host's IQNs
				newIQNs := []string{}
				for _, iqn := range host.IQNs {
					if !slices.Contains(hostPatch.RemoveIQNs, iqn) {
						newIQNs = append(newIQNs, iqn)
					}
				}
				host.IQNs = newIQNs
			}
			if len(hostPatch.AddIQNs) > 0 {
				host.IQNs = append(host.IQNs, hostPatch.AddIQNs...)
			}
			//NQN
			if len(hostPatch.NQNs) > 0 {
				host.NQNs = hostPatch.NQNs
			}
			if len(hostPatch.RemoveNQNs) > 0 {
				// Remove NQNs from the host's NQNs
				newNQNs := []string{}
				for _, nqn := range host.NQNs {
					if !slices.Contains(hostPatch.RemoveNQNs, nqn) {
						newNQNs = append(newNQNs, nqn)
					}
				}
				host.NQNs = newNQNs
			}
			if len(hostPatch.AddNQNs) > 0 {
				host.NQNs = append(host.NQNs, hostPatch.AddNQNs...)
			}
			//WWN
			if len(hostPatch.WWNs) > 0 {
				host.WWNs = hostPatch.WWNs
			}
			if len(hostPatch.RemoveWWNs) > 0 {
				// Remove WWNs from the host's WWNs
				newWWNs := []string{}
				for _, wwn := range host.WWNs {
					if !slices.Contains(hostPatch.RemoveWWNs, wwn) {
						newWWNs = append(newWWNs, wwn)
					}
				}
				host.WWNs = newWWNs
			}
			if len(hostPatch.AddWWNs) > 0 {
				host.WWNs = append(host.WWNs, hostPatch.AddWWNs...)
			}

			if hostPatch.CHAP != nil {
				host.Chap = *hostPatch.CHAP
			}

			if hostPatch.HostGroup != nil {
				host.HostGroup = *hostPatch.HostGroup
			}
			array.Hosts[i] = host
			return host, nil
		}
	}
	return nil, fmt.Errorf("host with name %s not found", name)
}

func (array *Array) DeleteHost(name string) error {
	for i, host := range array.Hosts {
		if host.Name == name {
			array.Hosts = append(array.Hosts[:i], array.Hosts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("host with name %s not found", name)
}
