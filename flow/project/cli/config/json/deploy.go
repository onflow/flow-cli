/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package json

import (
	"encoding/json"

	"github.com/onflow/flow-cli/flow/project/cli/config"
)

type jsonDeployments map[string]jsonDeploy

// transformToConfig transforms json structures to config structure
func (j jsonDeployments) transformToConfig() config.Deployments {
	deployments := make(config.Deployments, 0)

	for networkName, d := range j {
		for accountName, contracts := range d.Simple {
			deploy := config.Deploy{
				Network:   networkName,
				Account:   accountName,
				Contracts: contracts,
			}

			deployments = append(deployments, deploy)
		}
	}

	return deployments
}

// transformToJSON transforms config structure to json structures for saving
func transformDeploymentsToJSON(deployments config.Deployments) jsonDeployments {
	jsonDeployments := jsonDeployments{}

	for _, d := range deployments {
		if _, exists := jsonDeployments[d.Network]; exists {
			jsonDeployments[d.Network].Simple[d.Account] = d.Contracts
		} else {
			jsonDeployments[d.Network] = jsonDeploy{
				Simple: map[string][]string{
					d.Account: d.Contracts,
				},
			}
		}
	}

	return jsonDeployments
}

type Simple map[string][]string

type jsonDeploy struct {
	Simple
	// TODO: advanced format will include initializer arguments
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
