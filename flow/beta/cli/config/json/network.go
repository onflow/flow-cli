package json

import (
	"encoding/json"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
	"github.com/onflow/flow-go-sdk"
)

type jsonNetworks map[string]jsonNetwork

func (j jsonNetworks) transformToConfig() config.Networks {
	networks := make(config.Networks, 0)

	for networkName, n := range j {
		var network config.Network

		if n.Host != "" {
			network = config.Network{
				Name: networkName,
				Host: n.Host,
			}
		} else {
			network = config.Network{
				Name:    networkName,
				Host:    n.Advanced.Host,
				ChainID: flow.ChainID(n.Advanced.ChainID),
			}
		}

		networks = append(networks, network)
	}

	return networks
}

type Advanced struct {
	Host    string `json:"host"`
	ChainID string `json:"chain"`
}

type jsonNetwork struct {
	Host     string
	Advanced Advanced
}

func (j *jsonNetwork) UnmarshalJSON(b []byte) error {
	// simple
	var host string
	err := json.Unmarshal(b, &host)
	if err == nil {
		j.Host = host
		return nil
	}

	// advanced
	var advanced Advanced
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		j.Advanced = advanced
	} else {
		return err
	}

	return nil
}
