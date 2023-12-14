package json

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flowkit/config"
)

//	type dependency struct {
//		RemoteSource string `json:"remoteSource"`
//	}
//
// type jsonDependencies map[string]dependency
type jsonDependencies map[string]string

func (j jsonDependencies) transformToConfig() (config.Dependencies, error) {
	deps := make(config.Dependencies, 0)

	for dependencyName, dependencySource := range j {
		depNetwork, depAddress, depContractName, err := config.ParseRemoteSourceString(dependencySource)
		if err != nil {
			return nil, fmt.Errorf("error parsing remote source for dependency %s: %w", dependencyName, err)
		}

		dep := config.Dependency{
			Name: dependencyName,
			RemoteSource: config.RemoteSource{
				NetworkName:  depNetwork,
				Address:      flow.HexToAddress(depAddress),
				ContractName: depContractName,
			},
		}

		deps = append(deps, dep)
	}

	return deps, nil
}

func transformDependenciesToJSON(configDependencies config.Dependencies) jsonDependencies {
	jsonDeps := jsonDependencies{}

	for _, dep := range configDependencies {
		jsonDeps[dep.Name] = buildRemoteSourceString(dep.RemoteSource)
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
