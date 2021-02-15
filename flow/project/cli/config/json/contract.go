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

	"github.com/onflow/flow-cli/flow/project/cli/config"
)

// jsonContracts maping
type jsonContracts map[string]jsonContract

// transformToConfig transforms json structures to config structure
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
// transformToJSON transforms config structure to json structures for saving
func transformContractsToJSON(contracts config.Contracts) jsonContracts {
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
