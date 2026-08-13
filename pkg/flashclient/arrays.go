package flashclient

import (
	"fmt"
	"net/http"
)

type FixedReference struct {
	Name string `json:"name,omitempty"`
	Id   string `json:"id,omitempty"`
}

type NoIdReference struct {
	Name string `json:"name,omitempty"`
}

func (f FixedReference) GetId() string { return f.Id }

func (f FixedReference) GetName() string { return f.Name }

func (f NoIdReference) GetName() string { return f.Name }

type Array struct {
	Id                string            `json:"id"`
	Name              string            `json:"name"`
	Banner            string            `json:"banner"`
	Capacity          float64           `json:"capacity"`
	ConsoleLock       bool              `json:"console_lock_enabled"`
	Encryption        Encryption        `json:"encryption"`
	EradicationConfig EradicationConfig `json:"eradication_config"`
	IdleTimeout       int               `json:"idle_timeout"`
	NtpServers        []string          `json:"ntp_servers"`
	Os                string            `json:"os"`
	Parity            float64           `json:"parity"`
	SCSITimeout       int               `json:"scsi_timeout"`
	Space             Space             `json:"space"`
	Version           string            `json:"version"`
}

type ArrayShort struct {
	Id             string  `json:"id"`
	Name           string  `json:"name"`
	FrozenAt       int     `json:"frozen_at"`
	MediatorStatus string  `json:"mediator_status"`
	PreElected     bool    `json:"pre_elected"`
	Progress       float64 `json:"progress"`
	Status         string  `json:"status"`
}

type ArrayTiny struct {
	FixedReference
}

type Context struct {
	FixedReference
}

type Encryption struct {
	DataAtRest    DataAtRest `json:"data_at_rest"`
	ModuleVersion string     `json:"module_version"`
}

type DataAtRest struct {
	Algorithm string `json:"algorithm"`
	Enabled   bool   `json:"enabled"`
}

type EradicationConfig struct {
	EradicationDelay  int    `json:"eradication_delay"`
	ManualEradication string `json:"manual_eradication"`
}

func (fa *FAClient) GetArrays() ([]Array, error) {
	err := fa.RefreshSession()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	uri := "/arrays"
	result := Results[Array]{}
	res, err := fa.RestClient.R().
		SetResult(&result).
		Get(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to get arrays: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get arrays, got status code %d", res.StatusCode())
	}

	return result.Items, nil
}
