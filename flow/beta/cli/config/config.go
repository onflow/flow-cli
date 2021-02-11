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
	Contracts Contracts
	Networks  Networks
	Accounts  Accounts
	Deploys   Deploys
}

type Contracts []Contract
type Networks []Network
type Accounts []Account
type Deploys []Deploy

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

type KeyType string

const (
	KeyTypeHex       KeyType = "hex"        // Hex private key with in memory signer
	KeyTypeGoogleKMS KeyType = "google-kms" // Google KMS signer
	KeyTypeShell     KeyType = "shell"      // Exec out to a shell script
)

// GetForNetwork get all contracts by network
func (c *Contracts) GetForNetwork(network string) Contracts {
	return funk.Filter(c, func(c Contract) bool {
		return c.Network == network
	}).(Contracts)
}

// GetByNameAndNetwork get contract array for account and network
func (c *Contracts) GetByNameAndNetwork(name string, network string) Contract {
	contracts := funk.Filter(c, func(c Contract) bool {
		return c.Network == network && c.Name == name
	}).(Contracts)

	// if we don't find contract by name and network return default for name
	if len(contracts) == 0 {
		return c.GetByName(name)
	}

	return contracts[0]
}

// GetByName get contract by name
func (c *Contracts) GetByName(name string) Contract {
	return funk.Filter(c, func(c Contract) bool {
		return c.Name == name
	}).(Contracts)[0]
}

// GetByNetwork returns all contracts for specific network
func (c *Contracts) GetByNetwork(network string) Contracts {
	return funk.Filter(c, func(c Contract) bool {
		return c.Network == network || c.Network == "" // if network is not defined return for all set value
	}).(Contracts)
}

// GetAccountByName get account by name
func (a *Accounts) GetByName(name string) Account {
	return funk.Filter(a, func(a Account) bool {
		return a.Name == name
	}).(Accounts)[0]
}

// GetByAddress get account by address
func (a *Accounts) GetByAddress(address string) Account {
	return funk.Filter(a, func(a Account) bool {
		return a.Address.String() == address
	}).(Accounts)[0]
}

// GetByNetwork get all deploys by network
func (d *Deploys) GetByNetwork(network string) Deploys {
	return funk.Filter(d, func(d Deploy) bool {
		return d.Network == network
	}).(Deploys)
}

// GetByAccountAndNetwork get deploy by account and network
func (d *Deploys) GetByAccountAndNetwork(account string, network string) []Deploy {
	return funk.Filter(d, func(d Deploy) bool {
		return d.Account == account && d.Network == network
	}).(Deploys)
}

// GetByName get network by name
func (n *Networks) GetByName(name string) Network {
	return funk.Filter(n, func(n Network) bool {
		return n.Name == name
	}).(Networks)[0]
}
