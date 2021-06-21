/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

import "fmt"

// Contract defines the configuration for a Cadence contract.
type Contract struct {
	Name    string
	Source  string
	Network string
	Alias   string
}

type Contracts []Contract

// IsAlias checks if contract has an alias
func (c *Contract) IsAlias() bool {
	return c.Alias != ""
}

// GetByNameAndNetwork get contract array for account and network
func (c *Contracts) GetByNameAndNetwork(name string, network string) *Contract {
	for _, contract := range *c {
		if contract.Network == network && contract.Name == name {
			return &contract
		}
	}

	// if we don't find contract by name and network create a new contract
	// and replace only name and source with existing
	cName := c.GetByName(name)
	if cName == nil {
		return nil
	}

	return &Contract{
		Name:    cName.Name,
		Network: network,
		Source:  cName.Source,
	}
}

// GetByName get contract by name
func (c *Contracts) GetByName(name string) *Contract {
	for _, contract := range *c {
		if contract.Name == name {
			return &contract
		}
	}

	return nil
}

// GetByNetwork returns all contracts for specific network
func (c *Contracts) GetByNetwork(network string) Contracts {
	var contracts []Contract

	for _, contract := range *c {
		if contract.Network == network || contract.Network == "" {
			contracts = append(contracts, contract)
		}
	}

	return contracts
}

// AddOrUpdate add new or update if already present
func (c *Contracts) AddOrUpdate(name string, contract Contract) {
	for i, existingContract := range *c {
		if existingContract.Name == name &&
			existingContract.Network == contract.Network {
			(*c)[i] = contract
			return
		}
	}

	*c = append(*c, contract)
}

func (c *Contracts) Remove(name string) error {
	contract := c.GetByName(name)
	if contract == nil {
		return fmt.Errorf("contract named %s does not exist in configuration", name)
	}

	for i, contract := range *c {
		if contract.Name == name {
			*c = append((*c)[0:i], (*c)[i+1:]...) // remove item
		}
	}

	return nil
}
