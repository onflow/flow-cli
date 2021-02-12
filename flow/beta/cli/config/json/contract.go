package json

import (
	"encoding/json"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
)

// jsonContracts maping
type jsonContracts map[string]jsonContract

// transformToContracts transforms json values to config contracts
func (j jsonContracts) transformToConfig() config.Contracts {
	contracts := make(config.Contracts, 0)

	for contractName, c := range j {
		if c.Source != "" {
			contract := config.Contract{
				Name:   contractName,
				Source: c.Source,
			}

			contracts = append(contracts, contract)
		}

		for networkName, source := range c.SourcesByNetwork {
			contract := config.Contract{
				Name:    contractName,
				Source:  source,
				Network: networkName,
			}

			contracts = append(contracts, contract)
		}
	}

	return contracts
}

//REF: if we already loaded json from config no need to do this just return
func (j jsonContracts) transformToJSON(contracts config.Contracts) jsonContracts {
	jsonContracts := jsonContracts{}

	for _, c := range contracts {
		// if simple case
		if c.Network == "" {
			jsonContracts[c.Name] = jsonContract{
				Source: c.Source,
			}
		} else { // if advanced config
			// check if we already created for this name then add or create
			if _, exists := jsonContracts[c.Name]; exists {
				jsonContracts[c.Name].SourcesByNetwork[c.Network] = c.Source
			} else {
				jsonContracts[c.Name] = jsonContract{
					SourcesByNetwork: map[string]string{c.Network: c.Source},
				}
			}
		}
	}

	return jsonContracts
}

// jsonContract structure for json parsing
type jsonContract struct {
	Source           string
	SourcesByNetwork map[string]string
}

func (j *jsonContract) UnmarshalJSON(b []byte) error {
	var source string
	var sourcesByNetwork map[string]string

	// simple
	err := json.Unmarshal(b, &source)
	if err == nil {
		j.Source = source
		return nil
	}

	// advanced
	err = json.Unmarshal(b, &sourcesByNetwork)
	if err == nil {
		j.SourcesByNetwork = sourcesByNetwork
	} else {
		return err
	}

	return nil
}

func (j jsonContract) MarshalJSON() ([]byte, error) {
	if j.Source != "" {
		return json.Marshal(j.Source)
	} else {
		return json.Marshal(j.SourcesByNetwork)
	}
}
