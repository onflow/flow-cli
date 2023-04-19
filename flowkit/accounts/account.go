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

package accounts

import (
	"fmt"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/onflow/flow-cli/flowkit/config"
)

// Account is defined by an address and name and contains an AccountKey which can be used for signing.
type Account struct {
	Name    string
	Address flow.Address
	Key     AccountKey
}

func FromConfig(conf *config.Config) (Accounts, error) {
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

func ToConfig(accounts Accounts) config.Accounts {
	accountConfs := make([]config.Account, 0)

	for _, account := range accounts {
		accountConfs = append(accountConfs, toConfig(account))
	}

	return accountConfs
}

func fromConfig(account config.Account) (*Account, error) {
	key, err := keyFromConfig(account.Key)
	if err != nil {
		return nil, err
	}

	return &Account{
		Name:    account.Name,
		Address: account.Address,
		Key:     key,
	}, nil
}

func toConfig(account Account) config.Account {
	var key config.AccountKey
	if account.Key != nil {
		key = account.Key.ToConfig()
	}

	return config.Account{
		Name:    account.Name,
		Address: account.Address,
		Key:     key,
	}
}

func NewEmulatorAccount(sigAlgo crypto.SignatureAlgorithm, hashAlgo crypto.HashAlgorithm) (*Account, error) {
	seed, err := randomSeed(crypto.MinSeedLength)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate emulator service key: %v", err)
	}

	return &Account{
		Name:    config.DefaultEmulator.ServiceAccount,
		Address: flow.ServiceAddress(flow.Emulator),
		Key:     NewHexAccountKeyFromPrivateKey(0, hashAlgo, privateKey),
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
		if acc.Name == name {
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
		accNames = append(accNames, acc.Name)
	}
	return accNames
}

// ByAddress get an account by address.
func (a Accounts) ByAddress(address flow.Address) (*Account, error) {
	for i := range a {
		if a[i].Address == address {
			return &a[i], nil
		}
	}

	return nil, fmt.Errorf("could not find account with address %s in the configuration", address)
}

// ByName get an account by name or returns and error if no account found
func (a Accounts) ByName(name string) (*Account, error) {
	for i := range a {
		if a[i].Name == name {
			return &a[i], nil
		}
	}

	return nil, fmt.Errorf("could not find account with name %s in the configuration", name)
}

// AddOrUpdate add account if missing or updates if present.
func (a *Accounts) AddOrUpdate(account *Account) {
	for i, acc := range *a {
		if acc.Name == account.Name {
			(*a)[i] = acc
			return
		}
	}

	*a = append(*a, *account)
}
