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

	"github.com/onflow/flow-go-sdk"
)

// Contract defines the configuration for a Cadence contract.
type Contract struct {
	Name     string
	Location string
	Aliases  Aliases
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
			*c = append((*c)[0:i], (*c)[i+1:]...) // remove item
		}
	}

	return nil
}
