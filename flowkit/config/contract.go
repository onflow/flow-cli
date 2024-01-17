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

package config

import (
	"fmt"
	"path/filepath"

	"github.com/onflow/flow-go/fvm/systemcontracts"
	flowGo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/flow-go-sdk"
	"golang.org/x/exp/slices"
)

// Contract defines the configuration for a Cadence contract.
type Contract struct {
	Name         string
	Location     string
	Aliases      Aliases
	IsDependency bool
}

// Alias defines an existing pre-deployed contract address for specific network.
type Alias struct {
	Network string
	Address flow.Address
}

type Aliases []Alias

func (a *Aliases) ByNetwork(network string) *Alias {
	for _, alias := range *a {
		if alias.Network == network {
			return &alias
		}
	}

	return nil
}

func (a *Aliases) Add(network string, address flow.Address) {
	for _, alias := range *a {
		if alias.Network == network {
			return // already exists
		}
	}
	*a = append(*a, Alias{
		Network: network,
		Address: address,
	})
}

type Contracts []Contract

// IsAliased checks if contract has an alias.
func (c *Contract) IsAliased() bool {
	return len(c.Aliases) > 0
}

// ByName get contract by name or return an error if it doesn't exist.
func (c *Contracts) ByName(name string) (*Contract, error) {
	for i, contract := range *c {
		if contract.Name == name {
			return &(*c)[i], nil
		}
	}

	return nil, fmt.Errorf("contract %s does not exist", name)
}

// AddOrUpdate add new or update if already present.
func (c *Contracts) AddOrUpdate(contract Contract) {
	for i, existingContract := range *c {
		if existingContract.Name == contract.Name {
			(*c)[i] = contract
			return
		}
	}

	*c = append(*c, contract)
}

// Remove contract by its name.
func (c *Contracts) Remove(name string) error {
	if _, err := c.ByName(name); err != nil {
		return err
	}

	for i, contract := range *c {
		if contract.Name == name {
			*c = slices.Delete(*c, i, i+1)
		}
	}

	return nil
}

const (
	NetworkEmulator = "emulator"
	NetworkTestnet  = "testnet"
	NetworkMainnet  = "mainnet"
)

var networkToChainID = map[string]flowGo.ChainID{
	NetworkEmulator: flowGo.Emulator,
	NetworkTestnet:  flowGo.Testnet,
	NetworkMainnet:  flowGo.Mainnet,
}

func isCoreContract(networkName, contractName, contractAddress string) bool {
	sc := systemcontracts.SystemContractsForChain(networkToChainID[networkName])

	for _, coreContract := range sc.All() {
		if coreContract.Name == contractName && coreContract.Address.String() == contractAddress {
			return true
		}
	}

	return false
}

func getCoreContractByName(networkName, contractName string) *systemcontracts.SystemContract {
	sc := systemcontracts.SystemContractsForChain(networkToChainID[networkName])

	for i, coreContract := range sc.All() {
		if coreContract.Name == contractName {
			return &sc.All()[i]
		}
	}

	return nil
}

// AddDependencyAsContract adds a dependency as a contract if it doesn't already exist.
func (c *Contracts) AddDependencyAsContract(dependency Dependency, networkName string) {
	var aliases []Alias

	// If core contract found by name and address matches, then use all core contract aliases across networks
	if isCoreContract(networkName, dependency.RemoteSource.ContractName, dependency.RemoteSource.Address.String()) {
		for _, networkStr := range []string{NetworkEmulator, NetworkTestnet, NetworkMainnet} {
			coreContract := getCoreContractByName(networkStr, dependency.RemoteSource.ContractName)
			if coreContract != nil {
				aliases = append(aliases, Alias{
					Network: networkStr,
					Address: flow.HexToAddress(coreContract.Address.String()),
				})
			}
		}
	} else {
		aliases = append(aliases, dependency.Aliases...)
	}

	// If no core contract match, then use the address in remoteSource as alias
	if len(aliases) == 0 {
		aliases = append(aliases, Alias{
			Network: dependency.RemoteSource.NetworkName,
			Address: dependency.RemoteSource.Address,
		})
	}

	contract := Contract{
		Name:         dependency.Name,
		Location:     filepath.ToSlash(fmt.Sprintf("imports/%s/%s", dependency.RemoteSource.Address, dependency.RemoteSource.ContractName)),
		Aliases:      aliases,
		IsDependency: true,
	}

	if _, err := c.ByName(contract.Name); err != nil {
		c.AddOrUpdate(contract)
	}
}

func (c *Contracts) DependencyContractByName(name string) *Contract {
	for i, contract := range *c {
		if contract.Name == name && contract.IsDependency {
			return &(*c)[i]
		}
	}

	return nil
}
