/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package services

import (
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

var (
	serviceAddress = flow.HexToAddress("f8d6e0586b0a20c7")
	serviceName    = "emulator-account"
	sigAlgo        = crypto.ECDSA_P256
	hashAlgo       = crypto.SHA3_256
	pubKey, _      = crypto.DecodePublicKeyHex(sigAlgo, "858a7d978b25d61f348841a343f79131f4b9fab341dd8a476a6f4367c25510570bf69b795fc9c3d2b7191327d869bcf848508526a3c1cafd1af34f71c7765117")
	helloContract  = []byte(`
		pub contract Hello {
			pub let greeting: String
			init() {
				self.greeting = "Hello, World!"
			}
			pub fun hello(): String {
				return self.greeting
			}
		}
	`)
)

func TestAccounts(t *testing.T) {

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "hello.cdc", helloContract, 0644)
	af := afero.Afero{mockFS}

	proj, err := flowkit.Init(af, crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	mock := tests.DefaultMockGateway()
	accounts := NewAccounts(mock, proj, output.NewStdoutLogger(output.NoneLog))

	serviceAcc := proj.Accounts().ByName(serviceName)

	t.Run("Get an Account", func(t *testing.T) {
		account, err := accounts.Get(serviceAddress)

		mock.AssertFunctionsCalled(t, mock.GetAccount)
		assert.NoError(t, err)
		assert.Equal(t, account.Address, serviceAddress)
	})

	t.Run("Create an Account", func(t *testing.T) {
		newAddress := flow.HexToAddress("192440c99cb17282")

		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			assert.Equal(t, tx.FlowTransaction().Authorizers[0], serviceAddress)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewAccountCreateResult(newAddress), nil
		}

		compareAddress := serviceAddress
		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			assert.Equal(t, address, compareAddress)
			compareAddress = newAddress
			return tests.NewAccountWithAddress(address.String()), nil
		}

		a, err := accounts.Create(
			serviceAcc,
			[]crypto.PublicKey{pubKey},
			[]int{1000},
			sigAlgo,
			hashAlgo,
			nil,
		)

		mock.AssertFunctionsCalled(t, mock.SendSignedTransaction, mock.GetTransactionResult, mock.GetAccount)
		assert.NotNil(t, a)
		assert.NoError(t, err)
		assert.Equal(t, len(a.Address), 8)
	})

	t.Run("Create an Account with Contract", func(t *testing.T) {
		newAddress := flow.HexToAddress("192440c99cb17282")

		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			assert.Equal(t, tx.FlowTransaction().Authorizers[0], serviceAddress)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "acct.contracts.add"))

			return tests.NewTransaction(), nil
		}

		mock.GetTransactionResultMock = func(tx *flow.Transaction) (*flow.TransactionResult, error) {
			return tests.NewAccountCreateResult(newAddress), nil
		}

		mock.GetAccountMock = func(address flow.Address) (*flow.Account, error) {
			return tests.NewAccountWithAddress(address.String()), nil
		}

		a, err := accounts.Create(
			serviceAcc,
			[]crypto.PublicKey{pubKey},
			[]int{1000},
			sigAlgo,
			hashAlgo,
			[]string{"Hello:hello.cdc"},
		)

		mock.AssertFunctionsCalled(t, mock.SendSignedTransaction, mock.GetTransactionResult, mock.GetAccount)
		assert.NotNil(t, a)
		assert.NoError(t, err)
		assert.Equal(t, len(a.Address), 8)
	})

	t.Run("Contract Add for Account", func(t *testing.T) {
		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			return tests.NewTransaction(), nil
		}

		a, err := accounts.AddContract(serviceAcc, "Hello", helloContract, false)

		mock.AssertFunctionsCalled(t, mock.SendSignedTransaction)
		assert.NotNil(t, a)
		assert.NoError(t, err)
		assert.Equal(t, len(a.Address), 8)
	})

	t.Run("Contract Update for Account", func(t *testing.T) {
		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			assert.Equal(t, tx.FlowTransaction().Authorizers[0], serviceAddress)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.update__experimental"))

			return tests.NewTransaction(), nil
		}

		account, err := accounts.AddContract(serviceAcc, "Hello", helloContract, true)

		mock.AssertFunctionsCalled(t, mock.SendSignedTransaction)
		assert.NotNil(t, account)
		assert.Equal(t, len(account.Address), 8)
		assert.NoError(t, err)
	})

	t.Run("Contract Remove for Account", func(t *testing.T) {
		mock.SendSignedTransactionMock = func(tx *flowkit.Transaction) (*flow.Transaction, error) {
			assert.Equal(t, tx.FlowTransaction().Authorizers[0], serviceAddress)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.remove"))

			return tests.NewTransaction(), nil
		}

		account, err := accounts.RemoveContract(serviceAcc, "Hello")

		mock.AssertFunctionsCalled(t, mock.SendSignedTransaction)
		assert.NotNil(t, account)
		assert.Equal(t, len(account.Address), 8)
		assert.NoError(t, err)
	})

}
