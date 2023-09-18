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
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit/config"
)

type jsonContracts map[string]jsonContract

// transformToConfig transforms json structures to config structure.
func (j jsonContracts) transformToConfig() (config.Contracts, error) {
	contracts := make(config.Contracts, 0)

	for contractName, c := range j {
		if c.Simple != "" {
			contract := config.Contract{
				Name:     contractName,
				Location: c.Simple,
			}

			contracts = append(contracts, contract)
		} else {
			contract := config.Contract{
				Name:     contractName,
				Location: c.Advanced.Source,
			}
			for network, alias := range c.Advanced.Aliases {
				address := flow.HexToAddress(alias)
				if address == flow.EmptyAddress {
					return nil, fmt.Errorf("invalid alias address for a contract")
				}

				contract.Aliases.Add(network, address)
			}
			contracts = append(contracts, contract)
		}
	}

	return contracts, nil
}

// transformToJSON transforms config structure to json structures for saving.
func transformContractsToJSON(contracts config.Contracts) jsonContracts {
	jsonContracts := jsonContracts{}

	for _, c := range contracts {
		// if simple case
		if !c.IsAliased() {
			jsonContracts[c.Name] = jsonContract{
				Simple: filepath.ToSlash(c.Location),
			}
		} else { // if advanced config
			// check if we already created for this name then add or create
			aliases := make(map[string]string)
			for _, alias := range c.Aliases {
				aliases[alias.Network] = alias.Address.String()
			}

			jsonContracts[c.Name] = jsonContract{
				Advanced: jsonContractAdvanced{
					Source:  filepath.ToSlash(c.Location),
					Aliases: aliases,
				},
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
		j.Simple = filepath.FromSlash(source)
		return nil
	}

	// advanced
	err = json.Unmarshal(b, &advancedFormat)
	if err == nil {
		j.Advanced = advancedFormat
		j.Advanced.Source = filepath.FromSlash(j.Advanced.Source)
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

func (j jsonContract) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			{
				Ref: "#/$defs/jsonContractAdvanced",
			},
		},
		Definitions: map[string]*jsonschema.Schema{
			"jsonContractAdvanced": jsonschema.Reflect(jsonContractAdvanced{}),
		},
	}
}
