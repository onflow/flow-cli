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
	"path"

	"github.com/manifoldco/promptui"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/thoas/go-funk"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/config/json"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

// ReaderWriter is implemented by any value that has ReadFile and WriteFile
// and it is used to load and save files.
type ReaderWriter interface {
	ReadFile(source string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// Contract is a Cadence contract definition for a project.
type Contract struct {
	Name           string
	Source         string
	AccountAddress flow.Address
	AccountName    string
	Args           []cadence.Value
}

// State manages the state for a Flow project.
type State struct {
	conf         *config.Config
	confLoader   *config.Loader
	readerWriter ReaderWriter
	accounts     *Accounts
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
	p.conf.Accounts = accountsToConfig(*p.accounts, p.confLoader.AccountsFromFile())
	err := p.confLoader.Save(p.conf, path)

	// if we have defined accounts to be saved to an external file, iterate over them and save them separately
	for name, location := range p.confLoader.AccountsFromFile() {
		acc, _ := p.accounts.ByName(name)
		account := toConfig(*acc, nil)
		account.UseAdvanceFormat = true // in case where we save accounts to a separate file we use advance format even if default value

		c := config.Empty()
		c.Accounts.AddOrUpdate(name, account)
		err = p.confLoader.Save(c, location)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return fmt.Errorf("failed to save project configuration to: %s", path)
	}

	return nil
}

//Defines a Mainnet Standard Contract ( e.g Core Contracts, FungibleToken, NonFungibleToken )
type StandardContract struct {
	Address  flow.Address
	InfoLink string
}

func (p *State) CheckForStandardContractUsageOnMainnet() error {

	standardContracts := map[string]StandardContract{
		"FungibleToken": {
			Address:  flow.HexToAddress("0xf233dcee88fe0abe"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/fungible-token",
		},
		"FlowToken": {
			Address:  flow.HexToAddress("0x1654653399040a61"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/flow-token",
		},
		"FlowFees": {
			Address:  flow.HexToAddress("0xf919ee77447b7497"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/flow-fees",
		},
		"FlowServiceAccount": {
			Address:  flow.HexToAddress("0xe467b9dd11fa00df"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowStorageFees": {
			Address:  flow.HexToAddress("0xe467b9dd11fa00df"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/service-account",
		},
		"FlowIDTableStaking": {
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/staking-contract-reference",
		},
		"FlowEpoch": {
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowClusterQC": {
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"FlowDKG": {
			Address:  flow.HexToAddress("0x8624b52f9ddcd04a"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/epoch-contract-reference",
		},
		"NonFungibleToken": {
			Address:  flow.HexToAddress("0x1d7e57aa55817448"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/non-fungible-token",
		},
		"MetadataViews": {
			Address:  flow.HexToAddress("0x1d7e57aa55817448"),
			InfoLink: "https://developers.flow.com/flow/core-contracts/nft-metadata",
		},
	}

	contracts, err := p.DeploymentContractsByNetwork("mainnet")

	if err != nil {
		return err
	}
	for _, contract := range contracts {
		standardContract, ok := standardContracts[contract.Name]
		if ok {

			fmt.Printf("It seems like you are trying to deploy %s to Mainnet \n", contract.Name)
			fmt.Printf("It is a standard contract already deployed at address 0x%s \n", standardContract.Address.String())
			fmt.Printf("You can read more about it here: %s \n", standardContract.InfoLink)

			useMainnetVersionPrompt := promptui.Select{
				Label: "Do you wish to use Mainnet version instead? (y/n)",
				Items: []string{"Yes", "No"},
			}
			_, useMainnetVersion, err := useMainnetVersionPrompt.Run()
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}

			if useMainnetVersion == "Yes" {

				//replace contract with alias
				c, err := p.conf.Contracts.ByNameAndNetwork(contract.Name, "mainnet")
				if err != nil {
					return err
				}
				c.Alias = standardContract.Address.String()

				//remove from deploy
				for di, d := range p.conf.Deployments {
					if d.Network != "mainnet" {
						continue
					}
					for ci, c := range d.Contracts {
						if c.Name == contract.Name {
							p.conf.Deployments[di].Contracts = append((d.Contracts)[0:ci], (d.Contracts)[ci+1:]...)
							break
						}
					}
				}
			}
		}
	}

	return nil
}

// ContractConflictExists returns true if the same contract is configured to deploy
// to more than one account in the same network.
//
// The CLI currently does not allow the same contract to be deployed to multiple
// accounts in the same network.
func (p *State) ContractConflictExists(network string) bool {
	contracts, err := p.DeploymentContractsByNetwork(network)
	if err != nil {
		return false
	}

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
func (p *State) Accounts() *Accounts {
	return p.accounts
}

// Config get underlying configuration for advanced usage.
func (p *State) Config() *config.Config {
	return p.conf
}

// SetAccountFileLocation sets a private file location for the specified account.
func (p *State) SetAccountFileLocation(account Account, location string) {
	p.confLoader.SetAccountFromFile(account.name, location)
}

// EmulatorServiceAccount returns the service account for the default emulator profile.
func (p *State) EmulatorServiceAccount() (*Account, error) {
	emulator := p.conf.Emulators.Default()
	if emulator == nil {
		return nil, fmt.Errorf("no default emulator account")
	}

	return p.accounts.ByName(emulator.ServiceAccount)
}

// SetEmulatorKey sets the default emulator service account private key.
func (p *State) SetEmulatorKey(privateKey crypto.PrivateKey) {
	acc, _ := p.EmulatorServiceAccount()
	acc.SetKey(
		NewHexAccountKeyFromPrivateKey(
			acc.Key().Index(),
			acc.Key().HashAlgo(),
			privateKey,
		),
	)
}

// DeploymentContractsByNetwork returns all contracts for a network.
func (p *State) DeploymentContractsByNetwork(network string) ([]Contract, error) {
	contracts := make([]Contract, 0)

	// get deployments for the specified network
	for _, deploy := range p.conf.Deployments.ByNetwork(network) {
		account, err := p.accounts.ByName(deploy.Account)
		if err != nil {
			return nil, err
		}

		// go through each contract in this deployment
		for _, deploymentContract := range deploy.Contracts {
			c, err := p.conf.Contracts.ByNameAndNetwork(deploymentContract.Name, network)
			if err != nil {
				return nil, err
			}

			contract := Contract{
				Name:           c.Name,
				Source:         path.Clean(c.Source),
				AccountAddress: account.address,
				AccountName:    account.name,
				Args:           deploymentContract.Args,
			}

			contracts = append(contracts, contract)
		}
	}

	return contracts, nil
}

// AccountNamesForNetwork returns all configured account names for a network.
func (p *State) AccountNamesForNetwork(network string) []string {
	names := make([]string, 0)

	for _, account := range *p.accounts {
		if len(p.conf.Deployments.ByAccountAndNetwork(account.name, network)) > 0 {
			if !util.ContainsString(names, account.name) {
				names = append(names, account.name)
			}
		}
	}

	return names
}

type Aliases map[string]string

// AliasesForNetwork returns all deployment aliases for a network.
func (p *State) AliasesForNetwork(network string) Aliases {
	aliases := make(Aliases)

	// get all contracts for selected network and if any has an address as target make it an alias
	for _, contract := range p.conf.Contracts.ByNetwork(network) {
		if contract.IsAlias() {
			aliases[path.Clean(contract.Source)] = contract.Alias
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
	_, err = conf.Accounts.ByName(config.DefaultEmulatorServiceAccountName)
	if err == nil && len(conf.Emulators) == 0 {
		conf.Emulators.AddOrUpdate("", config.DefaultEmulator())
	}
	proj, err := newProject(conf, confLoader, readerWriter)
	if err != nil {
		return nil, fmt.Errorf("invalid project configuration: %s", err)
	}

	return proj, nil
}

// Exists checks if a project configuration exists.
func Exists(path string) bool {
	return config.Exists(path)
}

// Init initializes a new Flow project.
func Init(readerWriter ReaderWriter, sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) (*State, error) {
	emulatorServiceAccount, err := generateEmulatorServiceAccount(sigAlgo, hashAlgo)
	if err != nil {
		return nil, err
	}

	loader := config.NewLoader(readerWriter)
	loader.AddConfigParser(json.NewParser())

	return &State{
		confLoader:   loader,
		readerWriter: readerWriter,
		conf:         config.Default(),
		accounts:     &Accounts{*emulatorServiceAccount},
	}, nil
}

// newProject creates a new project from a configuration object.
func newProject(conf *config.Config, loader *config.Loader, readerWriter ReaderWriter) (*State, error) {
	accounts, err := accountsFromConfig(conf)
	if err != nil {
		return nil, err
	}

	return &State{
		conf:         conf,
		readerWriter: readerWriter,
		confLoader:   loader,
		accounts:     &accounts,
	}, nil
}
