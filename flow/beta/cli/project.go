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
	"errors"
	"fmt"
	"path"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
	"github.com/onflow/flow-cli/flow/beta/cli/keys"
)

type Contract struct {
	Name   string
	Source string
	Target flow.Address
}

/** ======================================================
Project struct defines configuration and accounts needed
*/
type Project struct {
	conf     *config.Config
	accounts []*Account
}

// todo discuss if this should be moved to config (i think config will be a
// central point for configuration for whole cli and should be reused)

// DefaultConfigPath path to default config file
const DefaultConfigPath = "flow.json"

// LoadProject parses config and create project instance
func LoadProject() *Project {
	conf, err := config.Load(DefaultConfigPath)
	if err != nil {
		if errors.Is(err, config.ErrDoesNotExist) {
			Exitf(
				1,
				"Project config file %s does not exist. Please initialize first\n",
				DefaultConfigPath,
			)
		}

		Exitf(1, "Failed to open project configuration in %s", DefaultConfigPath)

		return nil
	}

	proj, err := newProject(conf)
	if err != nil {
		// TODO: replace with a more detailed error message
		Exitf(1, "Invalid project configuration: %s", err)
	}

	return proj
}

// ProjectExists check if config exists
func ProjectExists() bool {
	return config.Exists(DefaultConfigPath)
}

// InitProject create project structure
func InitProject() *Project {
	serviceAccount, serviceAccountKey := generateEmulatorServiceAccount()

	return &Project{
		conf:     defaultConfig(serviceAccountKey),
		accounts: []*Account{serviceAccount},
	}
}

// todo this should be moved to configuration
// default values for emulator
const (
	DefaultEmulatorConfigProfileName  = "default"
	defaultEmulatorNetworkName        = "emulator"
	defaultEmulatorServiceAccountName = "emulator-service-account"
	defaultEmulatorPort               = 3569
	defaultEmulatorHost               = "127.0.0.1:3569"
)

// todo this should be moved to config - it refferences config all over
// defaultConfig initialize config with default values
func defaultConfig(serviceAccountKey *keys.HexAccountKey) *config.Config {
	return &config.Config{
		Emulator: map[string]config.EmulatorConfigProfile{
			DefaultEmulatorConfigProfileName: {
				Port: defaultEmulatorPort,
				ServiceKey: config.EmulatorServiceKey{
					PrivateKey: serviceAccountKey.PrivateKeyHex(),
					SigAlgo:    serviceAccountKey.SigAlgo(),
					HashAlgo:   serviceAccountKey.HashAlgo(),
				},
			},
		},
		Networks: map[string]config.Network{
			defaultEmulatorNetworkName: {
				Host:    defaultEmulatorHost,
				ChainID: flow.Emulator,
			},
		},
	}
}

// generateEmulatorServieAccount create service account used by emulator
func generateEmulatorServiceAccount() (*Account, *keys.HexAccountKey) {
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
	}, serviceAccountKey
}

// newProject creates new project based on configuration
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

// Host get project host
func (p *Project) Host(network string) string {
	return p.conf.Networks[network].Host
}

// EmulatorConfig get project emulator config
func (p *Project) EmulatorConfig(profile string) config.EmulatorConfigProfile {
	return p.conf.Emulator[profile]
}

// todo addapt this function - it takes network and it must return contract defined here as source (path), target (address) for network
func (p *Project) Contracts(network string) []Contract {
	contracts := make([]Contract, 0)

	for account, contracts := range p.conf.Deploy[network] {

	}

	for contractName, contractSource := range p.conf.Contracts {

		target := p.getTargetAddress(network)

		contract := Contract{
			Name:   contractName,
			Source: path.Clean(contractSource), // path to contract
			Target: target,                     // target account
		}

		contracts = append(contracts, contract)
	}

	return contracts
}

// AccountByAddress get account by address for project
func (p *Project) AccountByAddress(address flow.Address) *Account {
	for _, account := range p.accounts {
		if account.Address() == address {
			return account
		}
	}

	return nil
}

// todo should be moved to config
// Save save project to config file
func (p *Project) Save() {
	p.conf.Accounts = accountsToConfig(p.accounts)

	err := config.Save(p.conf, DefaultConfigPath)
	if err != nil {
		Exitf(1, "Failed to save project configuration to \"%s\"", DefaultConfigPath)
	}
}

// getTargetAddress
func (p *Project) getTargetAddress(target string) flow.Address {

	for _, account := range p.accounts {
		if account.name == target {
			return account.address
		}
	}

	return flow.HexToAddress(target[2:])
}

/** ======================================================
Account struct defines properties for account used in deployment
*/
type Account struct {
	name    string
	address flow.Address
	chainID flow.ChainID
	keys    []keys.AccountKey
}

// Address get address of account
func (a *Account) Address() flow.Address {
	return a.address
}

// DefaultKey gets first default key for account
func (a *Account) DefaultKey() keys.AccountKey {
	return a.keys[0]
}

// accountsFromConfig convert the config structure to account internal structure
func accountsFromConfig(conf *config.Config) ([]*Account, error) {
	accounts := make([]*Account, 0, len(conf.Accounts))

	for accountName, accountConf := range conf.Accounts {
		account, err := accountFromConfig(accountName, accountConf)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// todo think of refactoring this to config - is there a reason is not merged?
// accountFromConfig convert single account from config. Implement service addresses
func accountFromConfig(accountName string, accountConf config.Account) (*Account, error) {
	accountKeys := make([]keys.AccountKey, 0, len(accountConf.Keys))

	for _, key := range accountConf.Keys {
		accountKey, err := keys.NewAccountKey(key)
		if err != nil {
			return nil, err
		}

		accountKeys = append(accountKeys, accountKey)
	}

	var address flow.Address

	if accountConf.Address == "service" {
		address = flow.ServiceAddress(accountConf.ChainID)
	} else {
		address = flow.HexToAddress(accountConf.Address)
	}

	return &Account{
		name:    accountName,
		address: address,
		chainID: accountConf.ChainID,
		keys:    accountKeys,
	}, nil
}

// accountsToConfig convert account array to configuration accounts
func accountsToConfig(accounts []*Account) map[string]config.Account {
	accountConfs := make(map[string]config.Account)

	for _, account := range accounts {
		accountConfs[account.name] = accountToConfig(account)
	}

	return accountConfs
}

// accountToConfig convert single account to configuration accounts
func accountToConfig(account *Account) config.Account {
	var address string

	if account.address == flow.ServiceAddress(account.chainID) {
		address = "service"
	} else {
		address = fmt.Sprintf("0x%s", account.address.Hex())
	}

	keyConfigs := make([]config.AccountKey, 0, len(account.keys))

	for _, key := range account.keys {
		keyConfigs = append(keyConfigs, key.ToConfig())
	}

	return config.Account{
		Address: address,
		ChainID: account.chainID,
		Keys:    keyConfigs,
	}
}
