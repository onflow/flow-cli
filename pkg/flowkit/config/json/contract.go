/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

type jsonContracts map[string]jsonContract

// transformToConfig transforms json structures to config structure.
func (j jsonContracts) transformToConfig() (config.Contracts, error) {
	contracts := make(config.Contracts, 0)

	for contractName, c := range j {
		if c.Simple != "" {
			contract := config.Contract{
				Name:   contractName,
				Source: c.Simple,
			}

			contracts = append(contracts, contract)
		} else {
			for network, alias := range c.Advanced.Aliases {
				_, err := config.StringToAddress(alias)
				if err != nil {
					return nil, fmt.Errorf("invalid alias address for a contract")
				}

				contract := config.Contract{
					Name:    contractName,
					Source:  c.Advanced.Source,
					Network: network,
					Alias:   alias,
				}

				contracts = append(contracts, contract)
			}
		}
	}

	return contracts, nil
}

// transformToJSON transforms config structure to json structures for saving.
func transformContractsToJSON(contracts config.Contracts) jsonContracts {
	jsonContracts := jsonContracts{}

	for _, c := range contracts {
		// if simple case
		if c.Network == "" {
			jsonContracts[c.Name] = jsonContract{
				Simple: c.Source,
			}
		} else { // if advanced config
			// check if we already created for this name then add or create
			if _, exists := jsonContracts[c.Name]; exists && jsonContracts[c.Name].Advanced.Aliases != nil {
				jsonContracts[c.Name].Advanced.Aliases[c.Network] = c.Alias
			} else {
				jsonContracts[c.Name] = jsonContract{
					Advanced: jsonContractAdvanced{
						Source:  c.Source,
						Aliases: map[string]string{c.Network: c.Alias},
					},
				}
			}
		}
	}

	return jsonContracts
}

// jsonContractAdvanced for json parsing advanced config.
type jsonContractAdvanced struct {
	Source  string            `json:"source"`
	Aliases map[string]string `json:"aliases"`
}

// jsonContract structure for json parsing.
type jsonContract struct {
	Simple   string
	Advanced jsonContractAdvanced
}

func (j *jsonContract) UnmarshalJSON(b []byte) error {
	var source string
	var advancedFormat jsonContractAdvanced

	// simple
	err := json.Unmarshal(b, &source)
	if err == nil {
		j.Simple = source
		return nil
	}

	// advanced
	err = json.Unmarshal(b, &advancedFormat)
	if err == nil {
		j.Advanced = advancedFormat
	} else {
		return err
	}

	return nil
}

func (j jsonContract) MarshalJSON() ([]byte, error) {
	if j.Simple != "" {
		return json.Marshal(j.Simple)
	} else {
		return json.Marshal(j.Advanced)
	}
}
