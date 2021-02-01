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

	"github.com/onflow/flow-go-sdk"

	"github.com/onflow/flow-cli/flow/beta/cli/config"
	"github.com/onflow/flow-cli/flow/beta/cli/keys"
)

type Project struct {
	conf     *config.Config
	accounts []*Account
}

const defaultConfigPath = "flow.json"

func LoadProject() *Project {
	conf, err := config.Load(defaultConfigPath)
	if err != nil {
		if errors.Is(err, config.ErrDoesNotExist) {
			Exitf(
				1,
				"Project config file %s does not exist. Please initialize first\n",
				defaultConfigPath,
			)
		}

		Exitf(1, "Failed to open project configuration in %s\n", defaultConfigPath)

		return nil
	}

	proj, err := newProject(conf)
	if err != nil {
		// TODO: replace with a more detailed error message
		Exitf(1, "Invalid project configuration: %s\n", err)
	}

	return proj
}

func newProject(conf *config.Config) (*Project, error) {
	accounts, err := accountsFromConf(conf)
	if err != nil {
		return nil, err
	}

	return &Project{
		conf:     conf,
		accounts: accounts,
	}, nil
}

func (p *Project) Host(network string) string {
	return p.conf.Networks[network].Host
}

func (p *Project) EmulatorConfig(profile string) config.EmulatorConfigProfile {
	return p.conf.Emulator[profile]
}

func (p *Project) AccountByAddress(address flow.Address) *Account {
	for _, account := range p.accounts {
		if account.Address() == address {
			return account
		}
	}

	return nil
}

func (p *Project) getTargetAddress(target string) flow.Address {
	account, accountExists := p.conf.Accounts[target]
	if accountExists {
		return account.Address
	}

	return flow.HexToAddress(target[2:])
}

type Account struct {
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

func accountsFromConf(conf *config.Config) ([]*Account, error) {
	accounts := make([]*Account, 0, len(conf.Accounts))

	for _, accountConf := range conf.Accounts {
		account, err := accountFromConf(accountConf)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func accountFromConf(accountConf config.Account) (*Account, error) {
	accountKeys := make([]keys.AccountKey, 0, len(accountConf.Keys))

	for _, key := range accountConf.Keys {
		accountKey, err := keys.NewAccountKey(key)
		if err != nil {
			return nil, err
		}

		accountKeys = append(accountKeys, accountKey)
	}

	return &Account{
		address: accountConf.Address,
		chainID: accountConf.ChainID,
		keys:    accountKeys,
	}, nil
}
