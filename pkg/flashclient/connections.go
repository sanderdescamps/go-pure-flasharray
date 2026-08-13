package flashclient

import (
	"fmt"
	"net/http"
	"strings"
)

type ProtocolEndpoint struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Connection struct {
	Host             HostShort        `json:"host"`
	HostGroup        HostGroupShort   `json:"host_group"`
	Lun              *int64           `json:"lun"`
	NSID             *int64           `json:"nsid"`
	ProtocolEndpoint ProtocolEndpoint `json:"protocol_endpoint"`
	Volume           VolumeShort      `json:"volume"`
}

type ConnectionPostBody struct {
	Lun              *int64            `json:"lun,omitempty"`
	ProtocolEndpoint *ProtocolEndpoint `json:"protocol_endpoint,omitempty"`
}

func (fa *FAClient) GetConnections() ([]Connection, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/connections"
	result := Results[Connection]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get connections, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) CreateHostConnections(hostNames []string, volumeIds []string, post ConnectionPostBody) ([]Connection, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/connections"
	result := Results[Connection]{}
	res, err := fa.RestClient.R().
		SetBody(post).
		SetQueryParam("host_names", strings.Join(hostNames, ",")).
		SetQueryParam("volume_ids", strings.Join(volumeIds, ",")).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create connection, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) CreateHostGroupConnections(hostGroupNames []string, volumeIds []string, post ConnectionPostBody) ([]Connection, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/connections"
	result := Results[Connection]{}
	res, err := fa.RestClient.R().
		SetBody(post).
		SetQueryParam("host_group_names", strings.Join(hostGroupNames, ",")).
		SetQueryParam("volume_ids", strings.Join(volumeIds, ",")).
		SetResult(&result).
		Post(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create connection, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}

func (fa *FAClient) DeleteHostConnections(hostNames []string, volumeIds []string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/connections"
	res, err := fa.RestClient.R().
		SetQueryParam("host_names", strings.Join(hostNames, ",")).
		SetQueryParam("volume_ids", strings.Join(volumeIds, ",")).
		Delete(uri)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete connection, got status code %d", res.StatusCode())
	}

	return nil
}

func (fa *FAClient) DeleteHostGroupConnections(hostGroupNames []string, volumeIds []string) error {
	err := fa.RefreshSession()
	if err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/connections"
	res, err := fa.RestClient.R().
		SetQueryParam("host_group_names", strings.Join(hostGroupNames, ",")).
		SetQueryParam("volume_ids", strings.Join(volumeIds, ",")).
		Delete(uri)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete connection, got status code %d", res.StatusCode())
	}

	return nil
}
