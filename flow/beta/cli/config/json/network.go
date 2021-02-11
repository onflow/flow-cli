package json

import (
	"encoding/json"
	"errors"

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

type jsonNetwork struct {
	Host     string
	Advanced struct {
		Host    string `json:"host"`
		ChainID string `json:"chain"`
	}
}

func (j *jsonNetwork) UnmarshalJSON(b []byte) error {
	var val interface{}

	err := json.Unmarshal(b, &val)
	if err != nil {
		return err
	}

	switch typedVal := val.(type) {
	case string:
		// simple: host string for network
		j.Host = typedVal
	case map[string]string: //TODO: try changing j.Advanced
		json.Unmarshal(b, &j.Advanced)
	default:
		return errors.New("invalid network definition")
	}

	return nil
}
