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
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit/config"
)

type jsonDependencies map[string]jsonDependency

func (j jsonDependencies) transformToConfig() (config.Dependencies, error) {
	deps := make(config.Dependencies, 0)

	for dependencyName, dependency := range j {
		var dep config.Dependency

		if dependency.Simple != "" {
			depNetwork, depAddress, depContractName, err := config.ParseSourceString(dependency.Simple)
			if err != nil {
				return nil, fmt.Errorf("error parsing source for dependency %s: %w", dependencyName, err)
			}

			dep = config.Dependency{
				Name: dependencyName,
				Source: config.Source{
					NetworkName:  depNetwork,
					Address:      flow.HexToAddress(depAddress),
					ContractName: depContractName,
				},
			}
		} else {
			depNetwork, depAddress, depContractName, err := config.ParseSourceString(dependency.Extended.Source)
			if err != nil {
				return nil, fmt.Errorf("error parsing source for dependency %s: %w", dependencyName, err)
			}

			dep = config.Dependency{
				Name:    dependencyName,
				Version: dependency.Extended.Version,
				Source: config.Source{
					NetworkName:  depNetwork,
					Address:      flow.HexToAddress(depAddress),
					ContractName: depContractName,
				},
			}

			for network, alias := range dependency.Extended.Aliases {
				address := flow.HexToAddress(alias)
				if address == flow.EmptyAddress {
					return nil, fmt.Errorf("invalid alias address for a contract")
				}

				dep.Aliases.Add(network, address)
			}
		}

		deps = append(deps, dep)
	}

	return deps, nil
}

func transformDependenciesToJSON(configDependencies config.Dependencies, configContracts config.Contracts) jsonDependencies {
	jsonDeps := jsonDependencies{}

	for _, dep := range configDependencies {
		aliases := make(map[string]string)

		depContract := configContracts.DependencyContractByName(dep.Name)
		if depContract != nil {
			for _, alias := range depContract.Aliases {
				aliases[alias.Network] = alias.Address.String()
			}
		}

		jsonDeps[dep.Name] = jsonDependency{
			Extended: jsonDependencyExtended{
				Source:  buildSourceString(dep.Source),
				Version: dep.Version,
				Aliases: aliases,
			},
		}
	}

	return jsonDeps
}

func buildSourceString(source config.Source) string {
	var builder strings.Builder

	builder.WriteString(source.NetworkName)
	builder.WriteString("://")
	builder.WriteString(source.Address.String())
	builder.WriteString(".")
	builder.WriteString(source.ContractName)

	return builder.String()
}

// jsonDependencyExtended for json parsing advanced config.
type jsonDependencyExtended struct {
	Source  string            `json:"source"`
	Version string            `json:"version"`
	Aliases map[string]string `json:"aliases"`
}

// jsonDependency structure for json parsing.
type jsonDependency struct {
	Simple   string
	Extended jsonDependencyExtended
}

func (j *jsonDependency) UnmarshalJSON(b []byte) error {
	var source string
	var extendedFormat jsonDependencyExtended

	// simple
	err := json.Unmarshal(b, &source)
	if err == nil {
		j.Simple = source
		return nil
	}

	// advanced
	err = json.Unmarshal(b, &extendedFormat)
	if err == nil {
		j.Extended = extendedFormat
	} else {
		return err
	}

	return nil
}

func (j jsonDependency) MarshalJSON() ([]byte, error) {
	if j.Simple != "" {
		return json.Marshal(j.Simple)
	} else {
		return json.Marshal(j.Extended)
	}
}

func (j jsonDependency) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			{
				Ref: "#/$defs/jsonDependencyExtended",
			},
		},
		Definitions: map[string]*jsonschema.Schema{
			"jsonDependencyExtended": jsonschema.Reflect(jsonDependencyExtended{}),
		},
	}
}
