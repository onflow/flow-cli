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

package util

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2/accounts"
)

func Test_GetAccountsByNetworks(t *testing.T) {
	t.Run("Returns accounts for specified networks", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		testnetAddr := flow.HexToAddress("8efde57e98c557fa")  // Valid testnet address
		emulatorAddr := flow.HexToAddress("f8d6e0586b0a20c7") // Valid emulator address
		mainnetAddr := flow.HexToAddress("1654653399040a61")  // Valid mainnet address

		testnetAccount := &accounts.Account{
			Name:    "testnet-account",
			Address: testnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}
		emulatorAccount := &accounts.Account{
			Name:    "emulator-account",
			Address: emulatorAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}
		mainnetAccount := &accounts.Account{
			Name:    "mainnet-account",
			Address: mainnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}

		state.Accounts().AddOrUpdate(testnetAccount)
		state.Accounts().AddOrUpdate(emulatorAccount)
		state.Accounts().AddOrUpdate(mainnetAccount)

		result := GetAccountsByNetworks(state, []string{"testnet", "emulator"})

		assert.Len(t, result, 2)

		names := make([]string, len(result))
		for i, acc := range result {
			names[i] = acc.Name
		}
		assert.Contains(t, names, "testnet-account")
		assert.Contains(t, names, "emulator-account")
		assert.NotContains(t, names, "mainnet-account")
	})
}

func Test_GetTestnetAccounts(t *testing.T) {
	t.Run("Returns testnet accounts only", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		// testnet-valid address
		testnetAddr := flow.HexToAddress("8efde57e98c557fa")

		testnetAccount := &accounts.Account{
			Name:    "testnet-account",
			Address: testnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}

		state.Accounts().AddOrUpdate(testnetAccount)

		result := GetTestnetAccounts(state)

		found := false
		for _, acc := range result {
			if acc.Name == "testnet-account" {
				found = true
				assert.True(t, IsAddressValidForNetwork(acc.Address, "testnet"))
				break
			}
		}
		assert.True(t, found, "testnet-account should be found in results")
	})
}

func Test_ResolveAddressOrAccountNameForNetworks(t *testing.T) {
	t.Run("Resolves valid testnet hex address", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		address, err := ResolveAddressOrAccountNameForNetworks("8efde57e98c557fa", state, []string{"testnet", "emulator"})

		require.NoError(t, err)
		assert.Equal(t, "8efde57e98c557fa", address.String())
	})

	t.Run("Resolves valid emulator hex address", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		address, err := ResolveAddressOrAccountNameForNetworks("f8d6e0586b0a20c7", state, []string{"testnet", "emulator"})

		require.NoError(t, err)
		assert.Equal(t, "f8d6e0586b0a20c7", address.String())
	})

	t.Run("Resolves testnet account name", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		// Add a testnet account to state
		testnetAddr := flow.HexToAddress("8efde57e98c557fa")
		testnetAccount := &accounts.Account{
			Name:    "my-testnet-account",
			Address: testnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(testnetAccount)

		address, err := ResolveAddressOrAccountNameForNetworks("my-testnet-account", state, []string{"testnet", "emulator"})

		require.NoError(t, err)
		assert.Equal(t, testnetAddr, address)
	})

	t.Run("Resolves emulator account name", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		// Add an emulator account to state
		emulatorAddr := flow.HexToAddress("f8d6e0586b0a20c7")
		emulatorAccount := &accounts.Account{
			Name:    "my-emulator-account",
			Address: emulatorAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(emulatorAccount)

		address, err := ResolveAddressOrAccountNameForNetworks("my-emulator-account", state, []string{"testnet", "emulator"})

		require.NoError(t, err)
		assert.Equal(t, emulatorAddr, address)
	})

	t.Run("Fails with mainnet account name", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		// Add a mainnet account
		mainnetAddr := flow.HexToAddress("1654653399040a61")
		mainnetAccount := &accounts.Account{
			Name:    "mainnet-account",
			Address: mainnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(mainnetAccount)

		address, err := ResolveAddressOrAccountNameForNetworks("mainnet-account", state, []string{"testnet", "emulator"})

		assert.Equal(t, flow.EmptyAddress, address)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "testnet and emulator addresses")
	})

	t.Run("Fails with mainnet hex address", func(t *testing.T) {
		_, state, _ := TestMocks(t)

		address, err := ResolveAddressOrAccountNameForNetworks("1654653399040a61", state, []string{"testnet", "emulator"})

		assert.Equal(t, flow.EmptyAddress, address)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "testnet and emulator addresses")
	})
}
