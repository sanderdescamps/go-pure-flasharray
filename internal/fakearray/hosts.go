package fakearray

import (
	"fmt"
	"slices"

	"github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
)

func NewHostPost(name string, hostPost flashclient.HostPostBody) flashclient.Host {
	newHost := flashclient.Host{
		HostShort: flashclient.HostShort{
			NoIdReference: flashclient.NoIdReference{
				Name: name,
			},
		},
		Personality: flashclient.None,
		IQNs:        hostPost.IQNs,
		NQNs:        hostPost.NQNs,
		WWNs:        hostPost.WWNs,
		VLAN:        "",
	}

	if hostPost.Personality != nil {
		newHost.Personality = *hostPost.Personality
	}

	if hostPost.VLAN != nil {
		newHost.VLAN = *hostPost.VLAN
	}

	if hostPost.CHAP != nil {
		newHost.CHAP = *hostPost.CHAP
	}

	return newHost
}

func (array *Array) GetHosts() []flashclient.Host {
	hosts := make([]flashclient.Host, len(array.Hosts))
	for i, host := range array.Hosts {
		hosts[i] = *host
	}

	return hosts
}

func (array *Array) GetHost(name string) (*flashclient.Host, error) {
	for _, host := range array.Hosts {
		if host.Name == name {
			return host, nil
		}
	}
	return nil, fmt.Errorf("host with name %s not found", name)
}

func (array *Array) AddHost(host flashclient.Host) (*flashclient.Host, error) {
	for _, h := range array.Hosts {
		if h.Name == host.Name {
			return nil, fmt.Errorf("host with name %s already exists", host.Name)
		}
	}
	array.Hosts = append(array.Hosts, &host)
	return &host, nil
}

func (array *Array) UpdateHost(name string, hostPatch flashclient.HostPatchBody) (*flashclient.Host, error) {
	for i := range array.Hosts {
		if array.Hosts[i].Name == name {
			if hostPatch.Name != nil && *hostPatch.Name != "" {
				array.Hosts[i].Name = *hostPatch.Name
			}

			if hostPatch.Personality != nil {
				array.Hosts[i].Personality = *hostPatch.Personality
			}
			if hostPatch.VLAN != nil {
				array.Hosts[i].VLAN = *hostPatch.VLAN
			}

			//IQN
			if len(hostPatch.IQNs) > 0 {
				array.Hosts[i].IQNs = hostPatch.IQNs
			}
			if len(hostPatch.RemoveIQNs) > 0 {
				// Remove IQNs from the host's IQNs
				newIQNs := []string{}
				for _, iqn := range array.Hosts[i].IQNs {
					if !slices.Contains(hostPatch.RemoveIQNs, iqn) {
						newIQNs = append(newIQNs, iqn)
					}
				}
				array.Hosts[i].IQNs = newIQNs
			}
			if len(hostPatch.AddIQNs) > 0 {
				array.Hosts[i].IQNs = append(array.Hosts[i].IQNs, hostPatch.AddIQNs...)
			}
			//NQN
			if len(hostPatch.NQNs) > 0 {
				array.Hosts[i].NQNs = hostPatch.NQNs
			}
			if len(hostPatch.RemoveNQNs) > 0 {
				// Remove NQNs from the host's NQNs
				newNQNs := []string{}
				for _, nqn := range array.Hosts[i].NQNs {
					if !slices.Contains(hostPatch.RemoveNQNs, nqn) {
						newNQNs = append(newNQNs, nqn)
					}
				}
				array.Hosts[i].NQNs = newNQNs
			}
			if len(hostPatch.AddNQNs) > 0 {
				array.Hosts[i].NQNs = append(array.Hosts[i].NQNs, hostPatch.AddNQNs...)
			}
			//WWN
			if len(hostPatch.WWNs) > 0 {
				array.Hosts[i].WWNs = hostPatch.WWNs
			}
			if len(hostPatch.RemoveWWNs) > 0 {
				// Remove WWNs from the host's WWNs
				newWWNs := []string{}
				for _, wwn := range array.Hosts[i].WWNs {
					if !slices.Contains(hostPatch.RemoveWWNs, wwn) {
						newWWNs = append(newWWNs, wwn)
					}
				}
				array.Hosts[i].WWNs = newWWNs
			}
			if len(hostPatch.AddWWNs) > 0 {
				array.Hosts[i].WWNs = append(array.Hosts[i].WWNs, hostPatch.AddWWNs...)
			}

			if hostPatch.CHAP != nil {
				array.Hosts[i].CHAP = *hostPatch.CHAP
			}

			if hostPatch.HostGroup != nil {
				array.Hosts[i].HostGroup = *hostPatch.HostGroup
			}
			return array.Hosts[i], nil
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
