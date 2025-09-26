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

func Test_Fund(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Fail with invalid mainnet address", func(t *testing.T) {
		args := []string{"1654653399040a61"} // Mainnet address, not supported

		result, err := fund(
			args,
			command.GlobalFlags{},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "testnet and emulator addresses")
	})

	t.Run("Fail with mainnet account name", func(t *testing.T) {
		// Add a mainnet account to the state
		mainnetAddr := flow.HexToAddress("1654653399040a61")
		mainnetAccount := &accounts.Account{
			Name:    "mainnet-account",
			Address: mainnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(mainnetAccount)

		args := []string{"mainnet-account"} // Mainnet account name

		result, err := fund(
			args,
			command.GlobalFlags{},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "testnet and emulator addresses")
	})

	t.Run("Fail with no address and no fundable accounts", func(t *testing.T) {
		// Create state with only non-fundable account
		_, testState, _ := util.TestMocks(t)

		// Remove default account and add a non-fundable account
		_ = testState.Accounts().Remove("emulator-account")

		mainnetAddr := flow.HexToAddress("1654653399040a61")
		mainnetAccount := &accounts.Account{
			Name:    "mainnet-account",
			Address: mainnetAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		testState.Accounts().AddOrUpdate(mainnetAccount)

		args := []string{}

		result, err := fund(
			args,
			command.GlobalFlags{},
			util.NoLogger,
			srv.Mock,
			testState,
		)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "No fundable accounts found")
	})
}
