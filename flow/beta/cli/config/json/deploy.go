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

func (j jsonDeploys) transformToJSON(deploys config.Deploys) jsonDeploys {
	jsonDeploys := jsonDeploys{}

	for _, d := range deploys {
		if _, exists := jsonDeploys[d.Network]; exists {
			jsonDeploys[d.Network].Simple[d.Account] = d.Contracts
		} else {
			jsonDeploys[d.Network] = jsonDeploy{
				Simple: map[string][]string{
					d.Account: d.Contracts,
				},
			}
		}
	}

	return jsonDeploys
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

func (j jsonDeploy) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Simple)
}
