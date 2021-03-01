/*
 * Flow CLI
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/thoas/go-funk"
)

type Config struct {
	Emulators   Emulators
	Contracts   Contracts
	Networks    Networks
	Accounts    Accounts
	Deployments Deployments
}

type Contracts []Contract
type Networks []Network
type Accounts []Account
type Deployments []Deploy
type Emulators []Emulator

// Network config sets host and chain id
type Network struct {
	Name    string
	Host    string
	ChainID flow.ChainID
}

// Deploy structure for contract
type Deploy struct {
	Network   string   // network name to deploy to
	Account   string   // account name to which to deploy to
	Contracts []string // contracts names to deploy
}

// Contract is config for contract
type Contract struct {
	Name    string
	Source  string
	Network string
	Alias   string
}

// Account is main config for each account
type Account struct {
	Name    string
	Address flow.Address
	ChainID flow.ChainID
	Keys    []AccountKey
}

// AccountKey is config for account key
type AccountKey struct {
	Type     KeyType
	Index    int
	SigAlgo  crypto.SignatureAlgorithm
	HashAlgo crypto.HashAlgorithm
	Context  map[string]string
}

// Emulator is config for emulator
type Emulator struct {
	Name           string
	Port           int
	ServiceAccount string
}

type KeyType string

const (
	KeyTypeHex       KeyType = "hex"        // Hex private key with in memory signer
	KeyTypeGoogleKMS KeyType = "google-kms" // Google KMS signer
	KeyTypeShell     KeyType = "shell"      // Exec out to a shell script
)

// IsAlias checks if contract has an alias
func (c *Contract) IsAlias() bool {
	return c.Alias != ""
}

// GetByNameAndNetwork get contract array for account and network
func (c *Contracts) GetByNameAndNetwork(name string, network string) Contract {
	contracts := funk.Filter([]Contract(*c), func(c Contract) bool {
		return c.Network == network && c.Name == name
	}).([]Contract)

	// if we don't find contract by name and network create a new contract
	// and replace only name and source with existing
	if len(contracts) == 0 {
		cName := c.GetByName(name)

		return Contract{
			Name:    cName.Name,
			Network: network,
			Source:  cName.Source,
		}
	}

	return contracts[0]
}

// TODO: this filtering can cause error if not found, better to refactor to returning

// GetByName get contract by name
func (c *Contracts) GetByName(name string) Contract {
	return funk.Filter([]Contract(*c), func(c Contract) bool {
		return c.Name == name
	}).([]Contract)[0]
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

// GetAccountByName get account by name
func (a *Accounts) GetByName(name string) *Account {
	for _, account := range *a {
		if account.Name == name {
			return &account
		}
	}

	return nil
}

// GetByAddress get account by address
func (a *Accounts) GetByAddress(address string) *Account {
	for _, account := range *a {
		if account.Address.String() == address {
			return &account
		}
	}

	return nil
}

// GetByNetwork get all deployments by network
func (d *Deployments) GetByNetwork(network string) Deployments {
	var deployments []Deploy

	for _, deploy := range *d {
		if deploy.Network == network {
			deployments = append(deployments, deploy)
		}
	}

	return deployments
}

// GetByAccountAndNetwork get deploy by account and network
func (d *Deployments) GetByAccountAndNetwork(account string, network string) []Deploy {
	var deployments []Deploy

	for _, deploy := range *d {
		if deploy.Network == network && deploy.Account == account {
			deployments = append(deployments, deploy)
		}
	}

	return deployments
}

// GetByName get network by name
func (n *Networks) GetByName(name string) *Network {
	for _, network := range *n {
		if network.Name == name {
			return &network
		}
	}

	return nil
}

const DefaultEmulatorConfigName = "default"

// GetDefault emulator
func (e *Emulators) GetDefault() *Emulator {
	for _, emulator := range *e {
		if emulator.Name == DefaultEmulatorConfigName {
			return &emulator
		}
	}

	return nil
}
