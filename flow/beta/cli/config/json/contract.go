package json

import (
	"encoding/json"
	"errors"

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
	var val interface{}

	err := json.Unmarshal(b, &val)
	if err != nil {
		return err
	}

	switch typedVal := val.(type) {
	case string:
		// simple: single source for all networks
		j.Source = typedVal
	case map[string]string:
		// advanced: different source for each network
		j.SourcesByNetwork = typedVal
	default:
		return errors.New("invalid contract definition")
	}

	return nil
}
