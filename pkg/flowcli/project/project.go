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

	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/afero"
	"github.com/thoas/go-funk"

	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/config/json"
)

var DefaultConfigPath = "flow.json"

// Project has all the functionality to manage project
type Project struct {
	composer *config.Loader
	conf     *config.Config
	accounts []*Account
}

// Contract has all the functionality to manage contracts
type Contract struct {
	Name   string
	Source string
	Target flow.Address
}

// Load loads configuration and setup the project
func Load(configFilePath []string) (*Project, error) {
	composer := config.NewLoader(afero.NewOsFs())

	// here we add all available parsers (more to add yaml etc...)
	composer.AddConfigParser(json.NewParser())
	conf, err := composer.Load(configFilePath)

	if err != nil {
		if errors.Is(err, config.ErrDoesNotExist) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to open project configuration: %s", configFilePath)
	}

	proj, err := newProject(conf, composer)
	if err != nil {
		return nil, fmt.Errorf("invalid project configuration: %s", err)
	}

	return proj, nil
}

// Save saves project to path configuration
func (p *Project) Save(path string) error {
	p.conf.Accounts = accountsToConfig(p.accounts)
	err := p.composer.Save(p.conf, path)

	if err != nil {
		return fmt.Errorf("failed to save project configuration to: %s", path)
	}

	return nil
}

// Exists checks if project exists
func Exists(path string) bool {
	return config.Exists(path)
}

// Init initializes the project
func Init(sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) (*Project, error) {
	emulatorServiceAccount, err := generateEmulatorServiceAccount(sigAlgo, hashAlgo)

	composer := config.NewLoader(afero.NewOsFs())
	composer.AddConfigParser(json.NewParser())

	return &Project{
		composer: composer,
		conf:     defaultConfig(emulatorServiceAccount),
		accounts: []*Account{emulatorServiceAccount},
	}, err
}

const (
	DefaultEmulatorNetworkName        = "emulator"
	DefaultEmulatorServiceAccountName = "emulator-account"
	DefaultEmulatorPort               = 3569
	DefaultEmulatorHost               = "127.0.0.1:3569"
)

// defaultConfig creates new default configuration
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

// newProject creates new project from configuration passed
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

// CheckContractConflict checks if there is any contract duplication between accounts
// for now we don't allow two different accounts deploying same contract
func (p *Project) ContractConflictExists(network string) bool {
	contracts := p.GetContractsByNetwork(network)

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

// GetNetworkByName returns a network by name
func (p *Project) GetNetworkByName(name string) *config.Network {
	return p.conf.Networks.GetByName(name)
}

// EmulatorServiceAccount gets a service account for emulator
func (p *Project) EmulatorServiceAccount() (*Account, error) {
	emulator := p.conf.Emulators.GetDefault()
	acc := p.conf.Accounts.GetByName(emulator.ServiceAccount)
	return AccountFromConfig(*acc)
}

// SetEmulatorServiceKey sets emulator service key
func (p *Project) SetEmulatorServiceKey(privateKey crypto.PrivateKey) {
	acc := p.GetAccountByName(DefaultEmulatorServiceAccountName)
	acc.SetDefaultKey(
		NewHexAccountKeyFromPrivateKey(
			acc.DefaultKey().Index(),
			acc.DefaultKey().HashAlgo(),
			privateKey,
		),
	)
}

// GetContractsByNetwork return all contract for network
func (p *Project) GetContractsByNetwork(network string) []Contract {
	contracts := make([]Contract, 0)

	// get deployments for specific network
	for _, deploy := range p.conf.Deployments.GetByNetwork(network) {
		account := p.GetAccountByName(deploy.Account)

		// go through each contract for this deploy
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

// GetAllAccountNames gets all account names
func (p *Project) GetAllAccountNames() []string {
	names := make([]string, 0)

	for _, account := range p.accounts {
		if !util.StringContains(names, account.name) {
			names = append(names, account.name)
		}
	}

	return names
}

// GetAccountByName returns account by name
func (p *Project) GetAccountByName(name string) *Account {
	var account *Account

	for _, acc := range p.accounts {
		if acc.name == name {
			account = acc
		}
	}

	return account
}

// AddAccount adds account
func (p *Project) AddAccount(account *Account) {
	p.accounts = append(p.accounts, account)
}

// AddOrUpdateAccount addds or updates account
func (p *Project) AddOrUpdateAccount(account *Account) {
	for i, existingAccount := range p.accounts {
		if existingAccount.name == account.name {
			(*p).accounts[i] = account
			return
		}
	}

	p.accounts = append(p.accounts, account)
}

// GetAccountByAddress adds new account by address
func (p *Project) GetAccountByAddress(address string) *Account {
	for _, account := range p.accounts {
		if account.address.String() == flow.HexToAddress(address).String() {
			return account
		}
	}

	return nil
}

// GetAliases gets all deployment aliases for network
func (p *Project) GetAliases(network string) map[string]string {
	aliases := make(map[string]string)

	// get all contracts for selected network and if any has an address as target make it an alias
	for _, contract := range p.conf.Contracts.GetByNetwork(network) {
		if contract.IsAlias() {
			aliases[path.Clean(contract.Source)] = contract.Alias
		}
	}

	return aliases
}
