package json

import (
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"strings"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit/config"
)

type jsonDependencies map[string]jsonDependency

func (j jsonDependencies) transformToConfig() (config.Dependencies, error) {
	deps := make(config.Dependencies, 0)

	for dependencyName, dependency := range j {
		var dep config.Dependency

		if dependency.Simple != "" {
			depNetwork, depAddress, depContractName, err := config.ParseRemoteSourceString(dependency.Simple)
			if err != nil {
				return nil, fmt.Errorf("error parsing remote source for dependency %s: %w", dependencyName, err)
			}

			dep = config.Dependency{
				Name: dependencyName,
				RemoteSource: config.RemoteSource{
					NetworkName:  depNetwork,
					Address:      flow.HexToAddress(depAddress),
					ContractName: depContractName,
				},
			}
		} else {
			depNetwork, depAddress, depContractName, err := config.ParseRemoteSourceString(dependency.Extended.RemoteSource)
			if err != nil {
				return nil, fmt.Errorf("error parsing remote source for dependency %s: %w", dependencyName, err)
			}

			dep = config.Dependency{
				Name: dependencyName,
				RemoteSource: config.RemoteSource{
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
				RemoteSource: buildRemoteSourceString(dep.RemoteSource),
				Aliases:      aliases,
			},
		}
	}

	return jsonDeps
}

func buildRemoteSourceString(remoteSource config.RemoteSource) string {
	var builder strings.Builder

	builder.WriteString(remoteSource.NetworkName)
	builder.WriteString("://")
	builder.WriteString(remoteSource.Address.String())
	builder.WriteString(".")
	builder.WriteString(remoteSource.ContractName)

	return builder.String()
}

// jsonDependencyExtended for json parsing advanced config.
type jsonDependencyExtended struct {
	RemoteSource string            `json:"remoteSource"`
	Aliases      map[string]string `json:"aliases"`
}

// jsonDependency structure for json parsing.
type jsonDependency struct {
	Simple   string
	Extended jsonDependencyExtended
}

func (j *jsonDependency) UnmarshalJSON(b []byte) error {
	var remoteSource string
	var extendedFormat jsonDependencyExtended

	// simple
	err := json.Unmarshal(b, &remoteSource)
	if err == nil {
		j.Simple = remoteSource
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
