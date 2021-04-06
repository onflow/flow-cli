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

package project

import (
	"errors"
	"fmt"
	"path"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/thoas/go-funk"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/config/json"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

var (
	DefaultConfigPaths = []string{"flow.json"}
	DefaultConfigPath  = DefaultConfigPaths[0]
)

// Project contains the configuration for a Flow project.
type Project struct {
	composer *config.Loader
	conf     *config.Config
	accounts []*Account
}

// Contract is a Cadence contract definition for a project.
type Contract struct {
	Name   string
	Source string
	Target flow.Address
}

// Load loads a project configuration and returns the resulting project.
func Load(configFilePath []string) (*Project, error) {
	composer := config.NewLoader(afero.NewOsFs())

	// here we add all available parsers (more to add yaml etc...)
	composer.AddConfigParser(json.NewParser())
	conf, err := composer.Load(configFilePath)

	if err != nil {
		if errors.Is(err, config.ErrDoesNotExist) {
			return nil, err
		}

		return nil, err
	}

	proj, err := newProject(conf, composer)
	if err != nil {
		return nil, fmt.Errorf("invalid project configuration: %s", err)
	}

	return proj, nil
}

// Save saves the project configuration to the given path.
func (p *Project) Save(path string) error {
	p.conf.Accounts = accountsToConfig(p.accounts)
	err := p.composer.Save(p.conf, path)

	if err != nil {
		return fmt.Errorf("failed to save project configuration to: %s", path)
	}

	return nil
}

// Exists checks if a project configuration exists.
func Exists(path string) bool {
	return config.Exists(path)
}

// Init initializes a new Flow project.
func Init(sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) (*Project, error) {
	emulatorServiceAccount, err := generateEmulatorServiceAccount(sigAlgo, hashAlgo)
	if err != nil {
		return nil, err
	}

	composer := config.NewLoader(afero.NewOsFs())
	composer.AddConfigParser(json.NewParser())

	return &Project{
		composer: composer,
		conf:     defaultConfig(emulatorServiceAccount),
		accounts: []*Account{emulatorServiceAccount},
	}, nil
}

const (
	DefaultEmulatorNetworkName        = "emulator"
	DefaultEmulatorServiceAccountName = "emulator-account"
	DefaultEmulatorPort               = 3569
	DefaultEmulatorHost               = "127.0.0.1:3569"
)

// defaultConfig returns a new default configuration object.
func defaultConfig(defaultEmulatorServiceAccount *Account) *config.Config {
	return &config.Config{
		Emulators: config.Emulators{{
			Name:           config.DefaultEmulatorConfigName,
			ServiceAccount: defaultEmulatorServiceAccount.name,
			Port:           DefaultEmulatorPort,
		}},
		Networks: config.Networks{{
			Name:    DefaultEmulatorNetworkName,
			Host:    DefaultEmulatorHost,
			ChainID: flow.Emulator,
		}},
	}
}

// newProject creates a new project from a configuration object.
func newProject(conf *config.Config, composer *config.Loader) (*Project, error) {
	accounts, err := accountsFromConfig(conf)
	if err != nil {
		return nil, err
	}

	return &Project{
		composer: composer,
		conf:     conf,
		accounts: accounts,
	}, nil
}

// CheckContractConflict returns true if the same contract is configured to deploy
// to more than one account in the same network.
//
// The CLI currently does not allow the same contract to be deployed to multiple
// accounts in the same network.
func (p *Project) ContractConflictExists(network string) bool {
	contracts := p.ContractsByNetwork(network)

	uniq := funk.Uniq(
		funk.Map(contracts, func(c Contract) string {
			return c.Name
		}).([]string),
	).([]string)

	all := funk.Map(contracts, func(c Contract) string {
		return c.Name
	}).([]string)

	return len(all) != len(uniq)
}

// NetworkByName returns a network by name.
func (p *Project) NetworkByName(name string) *config.Network {
	return p.conf.Networks.GetByName(name)
}

// EmulatorServiceAccount returns the service account for the default emulator profilee.
func (p *Project) EmulatorServiceAccount() (*Account, error) {
	emulator := p.conf.Emulators.Default()
	acc := p.conf.Accounts.GetByName(emulator.ServiceAccount)
	return AccountFromConfig(*acc)
}

// SetEmulatorServiceKey sets the default emulator service account private key.
func (p *Project) SetEmulatorServiceKey(privateKey crypto.PrivateKey) {
	acc := p.AccountByName(DefaultEmulatorServiceAccountName)
	acc.SetDefaultKey(
		NewHexAccountKeyFromPrivateKey(
			acc.DefaultKey().Index(),
			acc.DefaultKey().HashAlgo(),
			privateKey,
		),
	)
}

// ContractsByNetwork returns all contracts for a network.
func (p *Project) ContractsByNetwork(network string) []Contract {
	contracts := make([]Contract, 0)

	// get deployments for the specified network
	for _, deploy := range p.conf.Deployments.GetByNetwork(network) {
		account := p.AccountByName(deploy.Account)

		// go through each contract in this deployment
		for _, contractName := range deploy.Contracts {
			c := p.conf.Contracts.GetByNameAndNetwork(contractName, network)

			contract := Contract{
				Name:   c.Name,
				Source: path.Clean(c.Source),
				Target: account.address,
			}

			contracts = append(contracts, contract)
		}
	}

	return contracts
}

// AllAccountName returns all configured account names.
func (p *Project) AllAccountName() []string {
	names := make([]string, 0)

	for _, account := range p.accounts {
		if !util.StringContains(names, account.name) {
			names = append(names, account.name)
		}
	}

	return names
}

// AddOrUpdateAccount adds or updates an account.
func (p *Project) AddOrUpdateAccount(account *Account) {
	for i, existingAccount := range p.accounts {
		if existingAccount.name == account.name {
			(*p).accounts[i] = account
			return
		}
	}

	p.accounts = append(p.accounts, account)
}

// AccountByAddress returns an account by address.
func (p *Project) AccountByAddress(address string) *Account {
	for _, account := range p.accounts {
		if account.address.String() == flow.HexToAddress(address).String() {
			return account
		}
	}

	return nil
}

// AccountByName returns an account by name.
func (p *Project) AccountByName(name string) *Account {
	var account *Account

	for _, acc := range p.accounts {
		if acc.name == name {
			account = acc
		}
	}

	return account
}

// AliasesForNetwork returns all deployment aliases for a network.
func (p *Project) AliasesForNetwork(network string) map[string]string {
	aliases := make(map[string]string)

	// get all contracts for selected network and if any has an address as target make it an alias
	for _, contract := range p.conf.Contracts.GetByNetwork(network) {
		if contract.IsAlias() {
			aliases[path.Clean(contract.Source)] = contract.Alias
		}
	}

	return aliases
}
