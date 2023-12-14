package config

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
)

func ParseRemoteSourceString(s string) (network, address, contractName string, err error) {
	parts := strings.Split(s, "://")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid dependency source format")
	}
	network = parts[0]

	subParts := strings.Split(parts[1], ".")
	if len(subParts) != 2 {
		return "", "", "", fmt.Errorf("invalid dependency source format")
	}
	address = subParts[0]
	contractName = subParts[1]

	return network, address, contractName, nil
}

type RemoteSource struct {
	NetworkName  string
	Address      flow.Address
	ContractName string
}

type Dependency struct {
	Name         string
	RemoteSource RemoteSource
}

type Dependencies []Dependency

func (d *Dependencies) ByName(name string) *Dependency {
	for i, dep := range *d {
		if dep.Name == name {
			return &(*d)[i]
		}
	}

	return nil
}

func (d *Dependencies) AddOrUpdate(dep Dependency) {
	for i, dependency := range *d {
		if dependency.Name == dep.Name {
			(*d)[i] = dep
			return
		}
	}

	*d = append(*d, dep)
}
