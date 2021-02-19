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

package cli

import (
	"path"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/thoas/go-funk"

	"github.com/onflow/flow-cli/flow/cli/keys"
	"github.com/onflow/flow-cli/flow/config"
	"github.com/onflow/flow-cli/flow/config/json"
)

const DefaultConfigPath = "flow.json"

type Project struct {
	conf     *config.Config
	accounts []*Account
}

func LoadProject() *Project {
	// TODO: dont have direct json loading here
	config, err := json.Load(DefaultConfigPath)
	if err != nil {
		Exitf(1, "Invalid project configuration: %s", err)
		return nil
	}

	proj, err := newProject(config)
	if err != nil {
		// TODO: replace with a more detailed error message
		Exitf(1, "Invalid project configuration: %s", err)
	}

	return proj
}

func ProjectExists() bool {
	return json.Exists()
}

func InitProject() *Project {
	emulatorServiceAccount := generateEmulatorServiceAccount()

	return &Project{
		conf:     defaultConfig(emulatorServiceAccount),
		accounts: []*Account{emulatorServiceAccount},
	}
}

const (
	defaultEmulatorNetworkName        = "emulator"
	defaultEmulatorServiceAccountName = "emulator-account"
	defaultEmulatorPort               = 3569
	defaultEmulatorHost               = "127.0.0.1:3569"
)

func defaultConfig(defaultEmulatorServiceAccount *Account) *config.Config {
	return &config.Config{
		Emulators: config.Emulators{{
			Name:           config.DefaultEmulatorConfigName,
			ServiceAccount: defaultEmulatorServiceAccount.name,
			Port:           defaultEmulatorPort,
		}},
		Networks: config.Networks{{
			Name:    defaultEmulatorNetworkName,
			Host:    defaultEmulatorHost,
			ChainID: flow.Emulator,
		}},
	}
}

func generateEmulatorServiceAccount() *Account {
	seed := RandomSeed(crypto.MinSeedLength)

	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	if err != nil {
		Exitf(1, "Failed to generate emulator service key: %v", err)
	}

	serviceAccountKey := keys.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, privateKey)

	return &Account{
		name:    defaultEmulatorServiceAccountName,
		address: flow.ServiceAddress(flow.Emulator),
		chainID: flow.Emulator,
		keys: []keys.AccountKey{
			serviceAccountKey,
		},
	}
}

func newProject(conf *config.Config) (*Project, error) {
	accounts, err := accountsFromConfig(conf)
	if err != nil {
		return nil, err
	}

	return &Project{
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

func (p *Project) HostWithOverride(host string) string {
	if host != "" {
		return host
	}
	// TODO fix this to support different networks (global flag)
	return p.conf.Networks.GetByName(config.DefaultEmulatorConfigName).Host
}

func (p *Project) Host(network string) string {
	return p.conf.Networks.GetByName(network).Host
}

func (p *Project) EmulatorServiceAccount() config.Account {
	emulator := p.conf.Emulators.GetDefault()
	return p.conf.Accounts.GetByName(emulator.ServiceAccount)
}

func (p *Project) GetContractsByNetwork(network string) []Contract {
	contracts := make([]Contract, 0)

	// get deploys for specific network
	for _, deploy := range p.conf.Deploys.GetByNetwork(network) {
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

func (p *Project) GetAllAccountNames() []string {
	return funk.Uniq(
		funk.Map(p.accounts, func(a *Account) string {
			return a.name
		}).([]string),
	).([]string)
}

func (p *Project) GetAccountByName(name string) *Account {
	return funk.Filter(p.accounts, func(a *Account) bool {
		return a.name == name
	}).([]*Account)[0]
}

func (p *Project) AddAccountByName(name string, account *Account) {

}

func (p *Project) GetAccountByAddress(address string) *Account {
	return funk.Filter(p.accounts, func(a *Account) bool {
		return a.address.String() == strings.ReplaceAll(address, "0x", "")
	}).([]*Account)[0]
}

func (p *Project) GetAliases(network string) map[string]string {
	aliases := make(map[string]string)

	// get all contracts for selected network and if any has an address as target make it an alias
	for _, contract := range p.conf.Contracts.GetByNetwork(network) {
		if contract.IsAlias() {
			aliases[contract.Name] = contract.Alias
		}
	}

	return aliases
}

func (p *Project) Save() {
	p.conf.Accounts = accountsToConfig(p.accounts)

	err := json.Save(p.conf, DefaultConfigPath)
	if err != nil {
		Exitf(1, "Failed to save project configuration to \"%s\"", DefaultConfigPath)
	}
}

type Contract struct {
	Name   string
	Source string
	Target flow.Address
}

type Account struct {
	name    string
	address flow.Address
	chainID flow.ChainID
	keys    []keys.AccountKey
}

func (a *Account) Address() flow.Address {
	return a.address
}

func (a *Account) DefaultKey() keys.AccountKey {
	return a.keys[0]
}

func accountsFromConfig(conf *config.Config) ([]*Account, error) {
	accounts := make([]*Account, 0, len(conf.Accounts))

	for _, accountConf := range conf.Accounts {
		account, err := accountFromConfig(accountConf)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func accountFromConfig(accountConf config.Account) (*Account, error) {
	accountKeys := make([]keys.AccountKey, 0, len(accountConf.Keys))

	for _, key := range accountConf.Keys {
		accountKey, err := keys.NewAccountKey(key)
		if err != nil {
			return nil, err
		}

		accountKeys = append(accountKeys, accountKey)
	}

	return &Account{
		name:    accountConf.Name,
		address: accountConf.Address,
		chainID: accountConf.ChainID,
		keys:    accountKeys,
	}, nil
}

func accountsToConfig(accounts []*Account) config.Accounts {
	accountConfs := make([]config.Account, 0)

	for _, account := range accounts {
		accountConfs = append(accountConfs, accountToConfig(account))
	}

	return accountConfs
}

func accountToConfig(account *Account) config.Account {
	keyConfigs := make([]config.AccountKey, 0, len(account.keys))

	for _, key := range account.keys {
		keyConfigs = append(keyConfigs, key.ToConfig())
	}

	return config.Account{
		Name:    account.name,
		Address: account.address,
		ChainID: account.chainID,
		Keys:    keyConfigs,
	}
}
