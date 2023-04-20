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
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestAccounts_ByName(t *testing.T) {
	acc1 := Account{
		Name:    "test1",
		Address: flow.HexToAddress("0x1"),
	}

	acc2 := Account{
		Name:    "test2",
		Address: flow.HexToAddress("0x2"),
	}

	accounts := Accounts{acc1, acc2}

	t.Run("Account present in slice", func(t *testing.T) {
		account, err := accounts.ByName("test1")
		assert.Nil(t, err)
		assert.Equal(t, &acc1, account)
	})

	t.Run("Account not present in slice", func(t *testing.T) {
		account, err := accounts.ByName("test3")
		assert.Error(t, err)
		assert.Nil(t, account)
	})
}

func TestAccounts_AddOrUpdate(t *testing.T) {
	acc1 := Account{
		Name:    "test1",
		Address: flow.HexToAddress("0x1"),
	}

	acc2 := Account{
		Name:    "test2",
		Address: flow.HexToAddress("0x2"),
	}

	// Test case 1: add a new account
	t.Run("Add new account", func(t *testing.T) {
		accounts := Accounts{acc1, acc2}
		acc3 := Account{
			Name:    "test3",
			Address: flow.HexToAddress("0x3"),
		}

		accounts.AddOrUpdate("test3", acc3)
		assert.Len(t, accounts, 3)
	})

	// Test case 2: update an existing account
	t.Run("Update existing account", func(t *testing.T) {
		accounts := Accounts{acc1, acc2}

		acc2Updated := Account{
			Name:    "test2",
			Address: flow.HexToAddress("0x4"),
		}

		accounts.AddOrUpdate("test2", acc2Updated)
		assert.Len(t, accounts, 2)

		account, err := accounts.ByName("test2")
		assert.Nil(t, err)
		assert.Equal(t, acc2Updated, *account)
	})
}

func TestAccounts_Remove(t *testing.T) {
	acc1 := Account{
		Name:    "account1",
		Address: flow.HexToAddress("01"),
	}
	acc2 := Account{
		Name:    "account2",
		Address: flow.HexToAddress("02"),
	}
	acc3 := Account{
		Name:    "account3",
		Address: flow.HexToAddress("03"),
	}

	accounts := Accounts{acc1, acc2, acc3}

	accounts.Remove("account2")
	assert.Equal(t, len(accounts), 2)

	_, err := accounts.ByName("account2")
	assert.Error(t, err)

	accounts.Remove("account4")
	assert.Equal(t, len(accounts), 2)
}
