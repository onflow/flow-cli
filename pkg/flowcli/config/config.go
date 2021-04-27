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

import (
	"errors"
	"fmt"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
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

// Network defines the configuration for a Flow network.
type Network struct {
	Name    string
	Host    string
	ChainID flow.ChainID
}

// Deploy defines the configuration for a contract deployment.
type Deploy struct {
	Network   string               // network name to deploy to
	Account   string               // account name to which to deploy to
	Contracts []ContractDeployment // contracts to deploy
}

// ContractDeployment defines the deployment of the contract with possible args
type ContractDeployment struct {
	Name string
	Args []ContractArgument
}

// ContractArgument defines contract init argument by name, type and value
type ContractArgument struct {
	Name string
	Arg  cadence.Value
}

// Contract defines the configuration for a Cadence contract.
type Contract struct {
	Name    string
	Source  string
	Network string
	Alias   string
}

// Account defines the configuration for a Flow account.
type Account struct {
	Name    string
	Address flow.Address
	ChainID flow.ChainID
	Key     AccountKey
}

// AccountKey defines the configuration for a Flow account key.
type AccountKey struct {
	Type     KeyType
	Index    int
	SigAlgo  crypto.SignatureAlgorithm
	HashAlgo crypto.HashAlgorithm
	Context  map[string]string
}

// Emulator defines the configuration for a Flow Emulator instance.
type Emulator struct {
	Name           string
	Port           int
	ServiceAccount string
}

type KeyType string

const (
	KeyTypeHex                        KeyType = "hex"        // Hex private key with in memory signer
	KeyTypeGoogleKMS                  KeyType = "google-kms" // Google KMS signer
	KeyTypeShell                      KeyType = "shell"      // Exec out to a shell script
	DefaultEmulatorConfigName                 = "default"
	PrivateKeyField                           = "privateKey"
	KMSContextField                           = "resourceName"
	DefaultEmulatorServiceAccountName         = "emulator-account"
)

// DefaultEmulatorNetwork get default emulator network
func DefaultEmulatorNetwork() Network {
	return Network{
		Name:    "emulator",
		Host:    "127.0.0.1:3569",
		ChainID: flow.Emulator,
	}
}

// DefaultTestnetNetwork get default testnet network
func DefaultTestnetNetwork() Network {
	return Network{
		Name:    "testnet",
		Host:    "access.devnet.nodes.onflow.org:9000",
		ChainID: flow.Testnet,
	}
}

// DefaultMainnetNetwork get default mainnet network
func DefaultMainnetNetwork() Network {
	return Network{
		Name:    "mainnet",
		Host:    "access.mainnet.nodes.onflow.org:9000",
		ChainID: flow.Mainnet,
	}
}

// DefaultNetworks gets all default networks
func DefaultNetworks() Networks {
	return Networks{
		DefaultEmulatorNetwork(),
		DefaultTestnetNetwork(),
		DefaultMainnetNetwork(),
	}
}

// DefaultEmulator gets default emulator
func DefaultEmulator() Emulator {
	return Emulator{
		Name:           "default",
		ServiceAccount: DefaultEmulatorServiceAccountName,
		Port:           3569,
	}
}

// DefaultConfig gets default configuration
func DefaultConfig() *Config {
	return &Config{
		Emulators: DefaultEmulators(),
		Networks:  DefaultNetworks(),
	}
}

// DefaultEmulators gets all default emulators
func DefaultEmulators() Emulators {
	return Emulators{DefaultEmulator()}
}

var ErrOutdatedFormat = errors.New("you are using old configuration format")

// IsAlias checks if contract has an alias
func (c *Contract) IsAlias() bool {
	return c.Alias != ""
}

// GetByNameAndNetwork get contract array for account and network
func (c *Contracts) GetByNameAndNetwork(name string, network string) *Contract {
	contracts := make(Contracts, 0)

	for _, contract := range *c {
		if contract.Network == network && contract.Name == name {
			contracts = append(contracts, contract)
		}
	}

	// if we don't find contract by name and network create a new contract
	// and replace only name and source with existing
	if len(contracts) == 0 {
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

	return &contracts[0]
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
		if existingContract.Name == name {
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

// AccountByName get account by name
func (a *Accounts) GetByName(name string) *Account {
	for _, account := range *a {
		if account.Name == name {
			return &account
		}
	}

	return nil
}

// AddOrUpdate add new or update if already present
func (a *Accounts) AddOrUpdate(name string, account Account) {
	for i, existingAccount := range *a {
		if existingAccount.Name == name {
			(*a)[i] = account
			return
		}
	}

	*a = append(*a, account)
}

// Remove remove account by name
func (a *Accounts) Remove(name string) error {
	account := a.GetByName(name)
	if account == nil {
		return fmt.Errorf("account named %s does not exist in configuration", name)
	}

	for i, account := range *a {
		if account.Name == name {
			*a = append((*a)[0:i], (*a)[i+1:]...) // remove item
		}
	}

	return nil
}

// GetByNetwork get all deployments by network
func (d *Deployments) GetByNetwork(network string) Deployments {
	var deployments Deployments

	for _, deploy := range *d {
		if deploy.Network == network {
			deployments = append(deployments, deploy)
		}
	}

	return deployments
}

// GetByAccountAndNetwork get deploy by account and network
func (d *Deployments) GetByAccountAndNetwork(account string, network string) Deployments {
	var deployments Deployments

	for _, deploy := range *d {
		if deploy.Network == network && deploy.Account == account {
			deployments = append(deployments, deploy)
		}
	}

	return deployments
}

// AddOrUpdate add new or update if already present
func (d *Deployments) AddOrUpdate(deployment Deploy) {
	for i, existingDeployment := range *d {
		if existingDeployment.Account == deployment.Account &&
			existingDeployment.Network == deployment.Network {
			(*d)[i] = deployment
			return
		}
	}

	*d = append(*d, deployment)
}

// Remove removes deployment by account and network
func (d *Deployments) Remove(account string, network string) error {
	deployment := d.GetByAccountAndNetwork(account, network)
	if deployment == nil {
		return fmt.Errorf(
			"deployment for account %s on network %s does not exist in configuration",
			account,
			network,
		)
	}

	for i, deployment := range *d {
		if deployment.Network == network && deployment.Account == account {
			*d = append((*d)[0:i], (*d)[i+1:]...) // remove item
		}
	}

	return nil
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

// AddOrUpdate add new network or update if already present
func (n *Networks) AddOrUpdate(name string, network Network) {
	for i, existingNetwork := range *n {
		if existingNetwork.Name == name {
			(*n)[i] = network
			return
		}
	}

	*n = append(*n, network)
}

func (n *Networks) Remove(name string) error {
	network := n.GetByName(name)
	if network == nil {
		return fmt.Errorf("network named %s does not exist in configuration", name)
	}

	for i, network := range *n {
		if network.Name == name {
			*n = append((*n)[0:i], (*n)[i+1:]...) // remove item
		}
	}

	return nil
}

// Default gets default emulator
func (e *Emulators) Default() *Emulator {
	for _, emulator := range *e {
		if emulator.Name == DefaultEmulatorConfigName {
			return &emulator
		}
	}

	return nil
}

// AddOrUpdate add new or update if already present
func (e *Emulators) AddOrUpdate(name string, emulator Emulator) {
	for i, existingEmulator := range *e {
		if existingEmulator.Name == name {
			(*e)[i] = emulator
			return
		}
	}

	*e = append(*e, emulator)
}
