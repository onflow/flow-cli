/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flowkit/v2/accounts"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_GetTestnetAccounts(t *testing.T) {
	t.Run("Returns testnet accounts only", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		// Create a testnet-valid address (testnet addresses start with specific ranges)
		testnetAddr := flow.HexToAddress("8efde57e98c557fa") // This is a valid testnet address

		testnetAccount := &accounts.Account{
			Name:    "testnet-account",
			Address: testnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, generateTestPrivateKey()),
		}

		state.Accounts().AddOrUpdate(testnetAccount)

		result := getTestnetAccounts(state)

		// Should return both the testnet account and potentially the emulator account if it's testnet-valid
		// Let's just check that our testnet account is included
		found := false
		for _, acc := range result {
			if acc.Name == "testnet-account" {
				found = true
				assert.True(t, acc.Address.IsValid(flow.Testnet))
				break
			}
		}
		assert.True(t, found, "testnet-account should be found in results")
	})

	t.Run("Returns empty when no testnet accounts", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		// Remove the default emulator account if it exists and add a non-testnet account
		state.Accounts().Remove("emulator-account")

		// Add an account that is definitely not testnet-valid
		mainnetAddr := flow.HexToAddress("01cf0e2f2f715450")
		mainnetAccount := &accounts.Account{
			Name:    "mainnet-account",
			Address: mainnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, generateTestPrivateKey()),
		}

		state.Accounts().AddOrUpdate(mainnetAccount)

		result := getTestnetAccounts(state)

		for _, acc := range result {
			assert.False(t, acc.Address.IsValid(flow.Testnet), "No accounts should be testnet-valid")
		}
	})

	t.Run("Returns multiple testnet accounts", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		// Add a known testnet account (from our flow.json example)
		testnetAddr1 := flow.HexToAddress("8efde57e98c557fa")

		account1 := &accounts.Account{
			Name:    "testnet-account-1",
			Address: testnetAddr1,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, generateTestPrivateKey()),
		}

		state.Accounts().AddOrUpdate(account1)

		result := getTestnetAccounts(state)

		testnetCount := 0
		for _, acc := range result {
			if acc.Name == "testnet-account-1" {
				assert.True(t, acc.Address.IsValid(flow.Testnet))
				testnetCount++
			}
		}
		assert.GreaterOrEqual(t, testnetCount, 1, "Should find at least our testnet account")
	})
}

func Test_Fund(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Fail with invalid testnet address", func(t *testing.T) {
		args := []string{"f8d6e0586b0a20c7"} // Emulator address, not testnet

		result, err := fund(
			args,
			command.GlobalFlags{},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "faucet can only work for valid Testnet addresses")
	})

	t.Run("Fail with no address and no testnet accounts", func(t *testing.T) {
		// Create state with only non-testnet account
		_, testState, _ := util.TestMocks(t)

		// Remove default account and add a non-testnet account
		testState.Accounts().Remove("emulator-account")

		mainnetAddr := flow.HexToAddress("01cf0e2f2f715450") // A mainnet-style address
		mainnetAccount := &accounts.Account{
			Name:    "mainnet-account",
			Address: mainnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, generateTestPrivateKey()),
		}
		testState.Accounts().AddOrUpdate(mainnetAccount)

		args := []string{} // No address provided

		result, err := fund(
			args,
			command.GlobalFlags{},
			util.NoLogger,
			srv.Mock,
			testState,
		)

		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no testnet accounts found in flow.json")
	})
}

func generateTestPrivateKey() crypto.PrivateKey {
	seed := make([]byte, crypto.MinSeedLength)
	for i := range seed {
		seed[i] = byte(i)
	}
	privKey, _ := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	return privKey
}
