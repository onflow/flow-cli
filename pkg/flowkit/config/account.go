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

package config

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

// Account defines the configuration for a Flow account.
type Account struct {
	Name    string
	Address flow.Address
	Key     AccountKey
}

type Accounts []Account

// AccountKey represents account key and all their possible configuration formats.
type AccountKey struct {
	Type       KeyType
	Index      int
	SigAlgo    crypto.SignatureAlgorithm
	HashAlgo   crypto.HashAlgorithm
	ResourceID string
	PrivateKey crypto.PrivateKey
}

// ByName get account by name.
func (a *Accounts) ByName(name string) (*Account, error) {
	for _, account := range *a {
		if account.Name == name {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account with name %s is not present in configuration", name)
}

// AddOrUpdate add new or update if already present.
func (a *Accounts) AddOrUpdate(name string, account Account) {
	for i, existingAccount := range *a {
		if existingAccount.Name == name {
			(*a)[i] = account
			return
		}
	}

	*a = append(*a, account)
}

// Remove remove account by name.
func (a *Accounts) Remove(name string) {
	for i, account := range *a {
		if account.Name == name {
			*a = append((*a)[0:i], (*a)[i+1:]...) // remove item
		}
	}
}
