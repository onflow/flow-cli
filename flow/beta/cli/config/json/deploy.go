package json

import (
	"encoding/json"
	"errors"

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

func (j jsonDeploy) UnmarshalJSON(b []byte) error {
	var val interface{}

	err := json.Unmarshal(b, &val)
	if err != nil {
		return err
	}

	switch typedVal := val.(type) {
	case Simple:
		j.Simple = typedVal
	default:
		return errors.New("invalid deploy definition")
	}

	return nil
}
