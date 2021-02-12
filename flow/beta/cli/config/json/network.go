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

func (j jsonNetworks) transformToJSON(networks config.Networks) jsonNetworks {
	jsonNetworks := jsonNetworks{}

	for _, n := range networks {
		// if simple case
		if n.ChainID == "" {
			jsonNetworks[n.Name] = jsonNetwork{
				Host: n.Host,
			}
		} else { // if advanced case
			jsonNetworks[n.Name] = jsonNetwork{
				Advanced: advanced{
					Host:    n.Host,
					ChainID: n.ChainID.String(),
				},
			}
		}
	}

	return jsonNetworks
}

type advanced struct {
	Host    string `json:"host"`
	ChainID string `json:"chain"`
}

type jsonNetwork struct {
	Host     string
	Advanced advanced
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
	var advanced advanced
	err = json.Unmarshal(b, &advanced)
	if err == nil {
		j.Advanced = advanced
	} else {
		return err
	}

	return nil
}

func (j jsonNetwork) MarshalJSON() ([]byte, error) {
	if j.Host != "" {
		return json.Marshal(j.Host)
	} else {
		return json.Marshal(j.Advanced)
	}
}
