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
