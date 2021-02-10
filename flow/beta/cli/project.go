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
	"path"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
	"github.com/onflow/flow-cli/flow/beta/cli/keys"
)

// Project structure containing current config
type Project struct {
	Config config.Config
}

// Contract structure defines Name of the contract, Source of
// the contract (path, url...), Target where to deploy contract to (account address)
type Contract struct {
	Name   string
	Source string
	Target flow.Address
}

// REF: this should be moved to config as it deals with config
// LoadProject parses config and create project instance
func LoadProject() *Project {
	configLocation := "flow.json" //TODO: support this from cli

	conf, err := config.Load(configLocation)
	if err != nil {
		if errors.Is(err, config.ErrDoesNotExist) {
			Exitf(
				1,
				"Project config file %s does not exist. Please initialize first\n",
				configLocation,
			)
		}

		Exitf(1, "Failed to open project configuration in %s", configLocation)

		return nil
	}

	return newProject(conf)
}

//REF: discuss and rethink what project init is and where configuration init should be stored
// InitProject create project structure
func InitProject() *Project {
	serviceAccount := generateEmulatorServiceAccount()

	config := config.Config{
		Accounts: config.AccountCollection{
			Accounts: map[string]config.Account{
				"emulator": *serviceAccount,
			},
		},
	}

	//TODO: create default config
	return newProject(&config)
}

func newProject(config *config.Config) *Project {
	return &Project{
		Config: *config,
	}
}

//REF: discuss and rethink what project init is and where configuration init should be stored
// default values for emulator
const (
	DefaultEmulatorConfigProfileName  = "default"
	defaultEmulatorNetworkName        = "emulator"
	defaultEmulatorServiceAccountName = "emulator-service-account"
	defaultEmulatorPort               = 3569
	defaultEmulatorHost               = "127.0.0.1:3569"
)

//REF: discuss and rethink what project init is and where configuration init should be stored
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
func generateEmulatorServiceAccount() *config.Account {
	seed := RandomSeed(crypto.MinSeedLength)

	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	if err != nil {
		Exitf(1, "Failed to generate emulator service key: %v", err)
	}

	serviceAccountKey := keys.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, privateKey)

	return &config.Account{
		Name:    defaultEmulatorServiceAccountName,
		Address: flow.ServiceAddress(flow.Emulator),
		ChainID: flow.Emulator,
		Keys: []config.AccountKey{{
			Type:     serviceAccountKey.Type(),
			Index:    serviceAccountKey.Index(),
			SigAlgo:  serviceAccountKey.SigAlgo(),
			HashAlgo: serviceAccountKey.HashAlgo(),
			Context:  serviceAccountKey.ToConfig().Context,
		}},
	}
}

func (p *Project) GetContractsByNetwork(network string) []Contract {
	contracts := make([]Contract, 0)

	// get deploys for specific network
	for _, deploy := range p.Config.Deploy.GetByNetwork(network) {
		account := p.Config.Accounts.GetByName(deploy.Account)
		// go through each contract for this deploy
		for _, contractName := range deploy.Contracts {
			c := p.Config.Contracts.GetByNameAndNetwork(contractName, network)

			contract := Contract{
				Name:   c.Name,
				Source: path.Clean(c.Source), //TODO: not necessary path - future improvements will include urls... REF: move this to config as validation and parsing
				Target: account.Address,
			}

			contracts = append(contracts, contract)
		}
	}

	return contracts
}

//TODO: should be moved to config
// Save save project to config file
func (p *Project) Save() {
	//TODO: implement this in current code
	/*
		p.conf.Accounts = accountsToConfig(p.accounts)

		err := config.Save(p.Config, DefaultConfigPath)
		if err != nil {
			Exitf(1, "Failed to save project configuration to \"%s\"", DefaultConfigPath)
		}*/
}
