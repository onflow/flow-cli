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

package flowkit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

	"github.com/onflow/flow-cli/flowkit/accounts"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/config/json"
	"github.com/onflow/flow-cli/flowkit/project"
)

// ReaderWriter defines read file and write file methods.
type ReaderWriter interface {
	ReadFile(source string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// State manages the state for a Flow project.
type State struct {
	conf         *config.Config
	confLoader   *config.Loader
	readerWriter ReaderWriter
	accounts     *accounts.Accounts
}

// ReaderWriter retrieve current file reader writer.
func (p *State) ReaderWriter() ReaderWriter {
	return p.readerWriter
}

// ReadFile exposes an injected file loader.
func (p *State) ReadFile(source string) ([]byte, error) {
	return p.readerWriter.ReadFile(source)
}

// SaveDefault saves to default path.
func (p *State) SaveDefault() error {
	return p.Save(config.DefaultPath)
}

// SaveEdited saves configuration to valid path.
func (p *State) SaveEdited(paths []string) error {
	// if paths are not default only allow specifying one config
	if !config.IsDefaultPath(paths) && len(paths) > 1 {
		return fmt.Errorf("specifying multiple paths is not supported when updating configuration")
	}
	// if default paths and local config doesn't exist don't allow updating global config
	if config.IsDefaultPath(paths) {
		_, err := p.confLoader.Load([]string{config.DefaultPath}) // check if default is present
		if err != nil {
			return fmt.Errorf("default configuration not found, please initialize it first or specify another configuration file")
		} else {
			return p.SaveDefault()
		}
	}

	return p.Save(paths[0])
}

// Save saves the project configuration to the given path.
func (p *State) Save(path string) error {
	p.conf.Accounts = accounts.ToConfig(*p.accounts)
	err := p.confLoader.Save(p.conf, path)

	if err != nil {
		return fmt.Errorf("failed to save project configuration to: %s", path)
	}

	return nil
}

// Networks get network configuration.
func (p *State) Networks() *config.Networks {
	return &p.conf.Networks
}

// Deployments get deployments configuration.
func (p *State) Deployments() *config.Deployments {
	return &p.conf.Deployments
}

// Contracts get contracts configuration.
func (p *State) Contracts() *config.Contracts {
	return &p.conf.Contracts
}

// Accounts get accounts.
func (p *State) Accounts() *accounts.Accounts {
	return p.accounts
}

// Config get underlying configuration for advanced usage.
func (p *State) Config() *config.Config {
	return p.conf
}

// EmulatorServiceAccount returns the service account for the default emulator profile.
func (p *State) EmulatorServiceAccount() (*accounts.Account, error) {
	emulator := p.conf.Emulators.Default()
	if emulator == nil {
		return nil, fmt.Errorf("no default emulator account")
	}

	return p.accounts.ByName(emulator.ServiceAccount)
}

// SetEmulatorKey sets the default emulator service account private key.
func (p *State) SetEmulatorKey(privateKey crypto.PrivateKey) {
	acc, _ := p.EmulatorServiceAccount()
	acc.Key = accounts.NewHexKeyFromPrivateKey(acc.Key.Index(), acc.Key.HashAlgo(), privateKey)
}

// DeploymentContractsByNetwork returns all contracts for a network.
//
// Build contract slice based on the network provided, check the deployment section for that network
// and retrieve the account by name, then add the accounts address on the contract as a destination.
func (p *State) DeploymentContractsByNetwork(network config.Network) ([]*project.Contract, error) {
	contracts := make([]*project.Contract, 0)

	// get deployments for the specified network
	for _, deploy := range p.conf.Deployments.ByNetwork(network.Name) {
		account, err := p.accounts.ByName(deploy.Account)
		if err != nil {
			return nil, err
		}

		// go through each contract in this deployment
		for _, deploymentContract := range deploy.Contracts {
			c, err := p.conf.Contracts.ByName(deploymentContract.Name)
			if err != nil {
				return nil, err
			}

			location := c.Location
			// if we loaded config from a single location, we should make the path of contracts defined in config relative to
			// config path we have provided, this will make cases where we execute loading in different path than config work
			if len(p.confLoader.LoadedLocations) == 1 {
				location = filepath.Join(
					filepath.Dir(p.confLoader.LoadedLocations[0]),
					location,
				)
			}

			code, err := p.readerWriter.ReadFile(location)
			if err != nil {
				return nil, errors.Wrap(err, "deployment by network failed to read contract code")
			}

			contract := project.NewContract(
				c.Name,
				filepath.Clean(location),
				code,
				account.Address,
				account.Name,
				deploymentContract.Args,
			)

			contracts = append(contracts, contract)
		}
	}

	return contracts, nil
}

// AccountsForNetwork returns all accounts used on a network defined by deployments.
func (p *State) AccountsForNetwork(network config.Network) *accounts.Accounts {
	exists := make(map[string]bool, 0)
	accs := make(accounts.Accounts, 0)

	for _, account := range *p.accounts {
		if p.conf.Deployments.ByAccountAndNetwork(account.Name, network.Name) != nil {
			slices.ContainsFunc(accs, func(a accounts.Account) bool {
				return a.Name == account.Name
			})
			if !exists[account.Name] {
				accs = append(accs, account)
			}
		}
	}
	return &accs
}

// AliasesForNetwork returns all deployment aliases for a network.
func (p *State) AliasesForNetwork(network config.Network) project.LocationAliases {
	aliases := make(project.LocationAliases)

	// get all contracts for selected network and if any has an address as target make it an alias
	for _, contract := range p.conf.Contracts {
		if contract.IsAliased() && contract.Aliases.ByNetwork(network.Name) != nil {
			alias := contract.Aliases.ByNetwork(network.Name).Address.String()
			aliases[filepath.Clean(contract.Location)] = alias // alias for import by file location
			aliases[contract.Name] = alias                 // alias for import by name
		}
	}

	return aliases
}

// Load loads a project configuration and returns the resulting project.
func Load(configFilePaths []string, readerWriter ReaderWriter) (*State, error) {
	confLoader := config.NewLoader(readerWriter)

	// here we add all available parsers (more to add yaml etc...)
	confLoader.AddConfigParser(json.NewParser())
	conf, err := confLoader.Load(configFilePaths)
	if err != nil {
		return nil, err
	}
	// only add a default emulator in the config if the emulator account is present in accounts
	_, err = conf.Accounts.ByName(config.DefaultEmulator.ServiceAccount)
	if err == nil && len(conf.Emulators) == 0 {
		conf.Emulators.AddOrUpdate("", config.DefaultEmulator)
	}
	proj, err := newProject(conf, confLoader, readerWriter)
	if err != nil {
		return nil, fmt.Errorf("invalid project configuration: %s", err)
	}

	return proj, nil
}

// Init initializes a new Flow project.
func Init(
	readerWriter ReaderWriter,
	sigAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) (*State, error) {
	emulatorServiceAccount, err := accounts.NewEmulatorAccount(sigAlgo, hashAlgo)
	if err != nil {
		return nil, err
	}

	loader := config.NewLoader(readerWriter)
	loader.AddConfigParser(json.NewParser())

	return &State{
		confLoader:   loader,
		readerWriter: readerWriter,
		conf:         config.Default(),
		accounts:     &accounts.Accounts{*emulatorServiceAccount},
	}, nil
}

// newProject creates a new project from a configuration object.
func newProject(
	conf *config.Config,
	loader *config.Loader,
	readerWriter ReaderWriter,
) (*State, error) {
	accs, err := accounts.FromConfig(conf)
	if err != nil {
		return nil, err
	}

	return &State{
		conf:         conf,
		readerWriter: readerWriter,
		confLoader:   loader,
		accounts:     &accs,
	}, nil
}
