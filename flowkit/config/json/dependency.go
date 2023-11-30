package json

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit/config"
)

type dependency struct {
	RemoteSource string            `json:"remoteSource"`
	Aliases      map[string]string `json:"aliases"`
}

type jsonDependencies map[string]dependency

func (j jsonDependencies) transformToConfig() (config.Dependencies, error) {
	deps := make(config.Dependencies, 0)

	for dependencyName, dependencies := range j {
		depNetwork, depAddress, depContractName, err := parseRemoteSourceString(dependencies.RemoteSource)
		if err != nil {
			return nil, fmt.Errorf("error parsing remote source for dependency %s: %w", dependencyName, err)
		}

		aliases := make(config.Aliases, 0)
		for network, address := range dependencies.Aliases {
			aliases = append(aliases, config.Alias{
				Network: network,
				Address: flow.HexToAddress(address),
			})
		}

		dep := config.Dependency{
			Name: dependencyName,
			RemoteSource: config.RemoteSource{
				NetworkName:  depNetwork,
				Address:      flow.HexToAddress(depAddress),
				ContractName: depContractName,
			},
			Aliases: aliases,
		}

		deps = append(deps, dep)
	}

	return deps, nil
}

func (j jsonDependencies) transformDependenciesToJSON(configDependencies config.Dependencies) jsonDependencies {
	jsonDeps := jsonDependencies{}

	for _, dep := range configDependencies {

		aliases := make(map[string]string)
		for _, alias := range dep.Aliases {
			aliases[alias.Network] = alias.Address.String()
		}

		jsonDeps[dep.Name] = dependency{
			RemoteSource: buildRemoteSourceString(dep.RemoteSource),
			Aliases:      aliases,
		}
	}

	return jsonDeps
}

func parseRemoteSourceString(s string) (network, address, contractName string, err error) {
	fmt.Printf("Parsing: %s\n", s)

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid format")
	}
	network = parts[0]

	subParts := strings.Split(parts[1], ".")
	if len(subParts) != 2 {
		return "", "", "", fmt.Errorf("invalid format")
	}
	address = subParts[0]
	contractName = subParts[1]

	return network, address, contractName, nil
}

func buildRemoteSourceString(remoteSource config.RemoteSource) string {
	var builder strings.Builder

	builder.WriteString(remoteSource.NetworkName)
	builder.WriteString("/")
	builder.WriteString(remoteSource.Address.String())
	builder.WriteString(".")
	builder.WriteString(remoteSource.ContractName)

	return builder.String()
}
