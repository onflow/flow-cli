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

package config

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_ExtractKey(t *testing.T) {
	t.Run("Success extracting key for specific account", func(t *testing.T) {
		srv, state, rw := util.TestMocks(t)

		testAddr := flow.HexToAddress("0x01cf0e2f2f715450")
		testAccount := &accounts.Account{
			Name:    "test-account",
			Address: testAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(testAccount)

		extractKeyFlags = flagsExtractKey{}

		result, err := extractKey(
			[]string{"test-account"},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "Successfully extracted keys for 1 account(s)")

		keyFilePath := accounts.PrivateKeyFile("test-account", "")
		keyData, err := rw.ReadFile(keyFilePath)
		require.NoError(t, err)
		assert.NotEmpty(t, keyData)

		updatedAccount, err := state.Accounts().ByName("test-account")
		require.NoError(t, err)
		assert.NotNil(t, updatedAccount)
	})

	t.Run("Success extracting keys for all accounts with --all flag", func(t *testing.T) {
		srv, state, rw := util.TestMocks(t)

		testAddr1 := flow.HexToAddress("0x01cf0e2f2f715450")
		testAccount1 := &accounts.Account{
			Name:    "test-account-1",
			Address: testAddr1,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(testAccount1)

		testAddr2 := flow.HexToAddress("0x179b6b1cb6755e31")
		testAccount2 := &accounts.Account{
			Name:    "test-account-2",
			Address: testAddr2,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(testAccount2)

		extractKeyFlags = flagsExtractKey{All: true}

		result, err := extractKey(
			[]string{},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "Successfully extracted keys for")

		keyFilePath1 := accounts.PrivateKeyFile("test-account-1", "")
		keyData1, err := rw.ReadFile(keyFilePath1)
		require.NoError(t, err)
		assert.NotEmpty(t, keyData1)

		keyFilePath2 := accounts.PrivateKeyFile("test-account-2", "")
		keyData2, err := rw.ReadFile(keyFilePath2)
		require.NoError(t, err)
		assert.NotEmpty(t, keyData2)

		extractKeyFlags = flagsExtractKey{}
	})

	t.Run("Fail when account not found", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)

		extractKeyFlags = flagsExtractKey{}

		result, err := extractKey(
			[]string{"nonexistent-account"},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("No accounts with inline keys", func(t *testing.T) {
		srv, state, rw := util.TestMocks(t)

		emulatorKeyFilePath := accounts.PrivateKeyFile("emulator-account", "")
		emulatorPrivateKey := util.GenerateTestPrivateKey()
		err := rw.WriteFile(emulatorKeyFilePath, []byte(emulatorPrivateKey.String()), 0600)
		require.NoError(t, err)

		testAddr := flow.HexToAddress("0x01cf0e2f2f715450")
		keyFilePath := accounts.PrivateKeyFile("file-key-account", "")

		privateKey := util.GenerateTestPrivateKey()
		err = rw.WriteFile(keyFilePath, []byte(privateKey.String()), 0600)
		require.NoError(t, err)

		testAccount := &accounts.Account{
			Name:    "file-key-account",
			Address: testAddr,
			Key:     accounts.NewFileKey(keyFilePath, 0, crypto.ECDSA_P256, crypto.SHA3_256, rw),
		}
		state.Accounts().AddOrUpdate(testAccount)

		extractKeyFlags = flagsExtractKey{All: true}

		result, err := extractKey(
			[]string{},
			command.GlobalFlags{ConfigPaths: []string{"flow.json"}},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "No accounts with inline keys found")

		extractKeyFlags = flagsExtractKey{}
	})
}

func Test_FindAccountsWithHexKeys(t *testing.T) {
	t.Run("Find accounts with hex keys", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		testAddr := flow.HexToAddress("0x01cf0e2f2f715450")
		testAccount := &accounts.Account{
			Name:    "test-hex-account",
			Address: testAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(testAccount)

		hexKeyAccounts := findAccountsWithHexKeys(state)

		assert.GreaterOrEqual(t, len(hexKeyAccounts), 1)
		assert.Contains(t, hexKeyAccounts, "test-hex-account")
	})

	t.Run("Skip accounts with file keys", func(t *testing.T) {
		_, state, rw := util.TestMocks(t)

		testAddr := flow.HexToAddress("0x01cf0e2f2f715450")
		keyFilePath := accounts.PrivateKeyFile("file-key-test", "")

		privateKey := util.GenerateTestPrivateKey()
		err := rw.WriteFile(keyFilePath, []byte(privateKey.String()), 0600)
		require.NoError(t, err)

		testAccount := &accounts.Account{
			Name:    "file-key-test",
			Address: testAddr,
			Key:     accounts.NewFileKey(keyFilePath, 0, crypto.ECDSA_P256, crypto.SHA3_256, rw),
		}
		state.Accounts().AddOrUpdate(testAccount)

		hexKeyAccounts := findAccountsWithHexKeys(state)

		assert.NotContains(t, hexKeyAccounts, "file-key-test")
	})
}

func Test_ExtractKeyForAccount(t *testing.T) {
	t.Run("Successfully extract key for account", func(t *testing.T) {
		_, state, rw := util.TestMocks(t)

		testAddr := flow.HexToAddress("0x01cf0e2f2f715450")
		privateKey := util.GenerateTestPrivateKey()
		testAccount := &accounts.Account{
			Name:    "extract-test",
			Address: testAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, privateKey),
		}
		state.Accounts().AddOrUpdate(testAccount)

		keyFilePath, err := extractKeyForAccount(state, "extract-test")

		require.NoError(t, err)
		assert.NotEmpty(t, keyFilePath)

		keyData, err := rw.ReadFile(keyFilePath)
		require.NoError(t, err)
		assert.Equal(t, privateKey.String(), string(keyData))

		updatedAccount, err := state.Accounts().ByName("extract-test")
		require.NoError(t, err)
		assert.NotNil(t, updatedAccount)
	})

	t.Run("Fail when account not found", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		_, err := extractKeyForAccount(state, "nonexistent")

		assert.ErrorContains(t, err, "not found")
	})

	t.Run("Fail when key file already exists", func(t *testing.T) {
		_, state, rw := util.TestMocks(t)

		testAddr := flow.HexToAddress("0x01cf0e2f2f715450")
		testAccount := &accounts.Account{
			Name:    "existing-file-test",
			Address: testAddr,
			Key:     accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, util.GenerateTestPrivateKey()),
		}
		state.Accounts().AddOrUpdate(testAccount)

		keyFilePath := accounts.PrivateKeyFile("existing-file-test", "")
		err := rw.WriteFile(keyFilePath, []byte("existing content"), 0600)
		require.NoError(t, err)

		_, err = extractKeyForAccount(state, "existing-file-test")

		assert.ErrorContains(t, err, "already exists")
	})
}
