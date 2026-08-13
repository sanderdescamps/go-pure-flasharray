package flashclient

import (
	"fmt"
	"net/http"
)

type Port struct {
	Name     string `json:"name"`
	Iqn      string `json:"iqn"`
	Nqn      string `json:"nqn"`
	Portal   string `json:"portal"`
	Wwn      string `json:"wwn"`
	Failover string `json:"failover"`
}

func (fa *FAClient) GetPorts() ([]Port, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/ports"
	result := Results[Port]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, err
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get ports, got status code %d", res.StatusCode())
	}
	return result.Items, nil
}
