package json

import (
	"encoding/json"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
)

type jsonDeploys map[string]jsonDeploy

func (j jsonDeploys) transformToConfig() config.Deploys {
	deploys := make(config.Deploys, 0)

	for networkName, d := range j {
		for accountName, contracts := range d.Simple {
			deploy := config.Deploy{
				Network:   networkName,
				Account:   accountName,
				Contracts: contracts,
			}

			deploys = append(deploys, deploy)
		}
	}

	return deploys
}

type Simple map[string][]string

type jsonDeploy struct {
	Simple
	//TODO: advanced format will include variables
}

func (j *jsonDeploy) UnmarshalJSON(b []byte) error {
	var simple map[string][]string

	err := json.Unmarshal(b, &simple)
	if err != nil {
		return err
	}
	j.Simple = simple

	return nil
}
