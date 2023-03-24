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
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

// todo refactor to config

// Account is a flowkit-specific account implementation.
type Account struct {
	name    string
	address flow.Address
	key     AccountKey
}

// NewAccount creates an empty account with the provided name.
func NewAccount(name string) *Account {
	return &Account{
		name: name,
	}
}

// Address get account address.
func (a *Account) Address() flow.Address {
	return a.address
}

// Name get account name.
func (a *Account) Name() string {
	return a.name
}

// Key get account key.
func (a *Account) Key() AccountKey {
	return a.key
}

// SetAddress sets the account address.
func (a *Account) SetAddress(address flow.Address) *Account {
	a.address = address
	return a
}

// SetName sets the account name.
func (a *Account) SetName(name string) *Account {
	a.name = name
	return a
}

// SetKey sets account key.
func (a *Account) SetKey(key AccountKey) *Account {
	a.key = key
	return a
}

func accountsFromConfig(conf *config.Config) (Accounts, error) {
	var accounts Accounts
	for _, accountConf := range conf.Accounts {
		acc, err := fromConfig(accountConf)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *acc)
	}

	return accounts, nil
}

func accountsToConfig(accounts Accounts) config.Accounts {
	accountConfs := make([]config.Account, 0)

	for _, account := range accounts {
		accountConfs = append(accountConfs, toConfig(account))
	}

	return accountConfs
}

func fromConfig(account config.Account) (*Account, error) {
	key, err := accountKeyFromConfig(account.Key)
	if err != nil {
		return nil, err
	}

	return &Account{
		name:    account.Name,
		address: account.Address,
		key:     key,
	}, nil
}

func toConfig(account Account) config.Account {
	var key config.AccountKey
	if account.key != nil {
		key = account.key.ToConfig()
	}

	return config.Account{
		Name:    account.name,
		Address: account.address,
		Key:     key,
	}
}

func generateEmulatorServiceAccount(sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) (*Account, error) {
	seed, err := util.RandomSeed(crypto.MinSeedLength)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate emulator service key: %v", err)
	}

	return &Account{
		name:    config.DefaultEmulatorServiceAccountName,
		address: flow.ServiceAddress(flow.Emulator),
		key:     NewHexAccountKeyFromPrivateKey(0, hashAlgo, privateKey),
	}, nil
}

// Accounts is a collection of account.
type Accounts []Account

// Remove an account.
func (a *Accounts) Remove(name string) error {
	account, err := a.ByName(name)
	if err != nil {
		return err
	}

	if account == nil {
		return fmt.Errorf("account named %s does not exist in configuration", name)
	}

	for i, acc := range *a {
		if acc.name == name {
			*a = append((*a)[0:i], (*a)[i+1:]...) // remove item
		}
	}

	return nil
}

func (a *Accounts) String() string {
	return strings.Join(a.Names(), ",")
}

func (a *Accounts) Names() []string {
	accNames := make([]string, 0)
	for _, acc := range *a {
		accNames = append(accNames, acc.name)
	}
	return accNames
}

// ByAddress get an account by address.
func (a *Accounts) ByAddress(address flow.Address) (*Account, error) {
	for _, acc := range *a {
		if acc.address == address {
			return &acc, nil
		}
	}

	return nil, fmt.Errorf("could not find account with address %s in the configuration", address)
}

// ByName get an account by name or returns and error if no account found
func (a *Accounts) ByName(name string) (*Account, error) {
	for _, acc := range *a {
		if acc.name == name {
			return &acc, nil
		}
	}
	return nil, fmt.Errorf("could not find account with name %s in the configuration", name)

}

// AddOrUpdate add account if missing or updates if present.
func (a *Accounts) AddOrUpdate(account *Account) {
	for i, acc := range *a {
		if acc.name == account.name {
			(*a)[i] = acc
			return
		}
	}

	*a = append(*a, *account)
}
