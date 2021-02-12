/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

	"github.com/onflow/flow-cli/flow/beta/cli/config"
)

type jsonDeploys map[string]jsonDeploy

// transformToConfig transforms json structures to config structure
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

// transformToJSON transforms config structure to json structures for saving
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
