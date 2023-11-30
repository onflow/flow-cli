package config

import "github.com/onflow/flow-go-sdk"

type RemoteSource struct {
	NetworkName  string
	Address      flow.Address
	ContractName string
}

type Dependency struct {
	Name         string
	RemoteSource RemoteSource
	Aliases      Aliases
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
