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

package services

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flowkit/gateway"

	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/pkg/flowkit/output"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/tests"
)

func setup() (*flowkit.State, *Services, *tests.TestGateway) {
	readerWriter := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	gw := tests.DefaultMockGateway()
	s := NewServices(gw.Mock, state, output.NewStdoutLogger(output.NoneLog))

	return state, s, gw
}

func TestAccounts(t *testing.T) {
	state, _, _ := setup()
	pubKey, _ := crypto.DecodePublicKeyHex(crypto.ECDSA_P256, "858a7d978b25d61f348841a343f79131f4b9fab341dd8a476a6f4367c25510570bf69b795fc9c3d2b7191327d869bcf848508526a3c1cafd1af34f71c7765117")
	serviceAcc, _ := state.EmulatorServiceAccount()
	serviceAddress := serviceAcc.Address()

	t.Run("Get an Account", func(t *testing.T) {
		_, s, gw := setup()
		account, err := s.Accounts.Get(serviceAddress)

		gw.Mock.AssertCalled(t, "GetAccount", serviceAddress)
		assert.NoError(t, err)
		assert.Equal(t, account.Address, serviceAddress)
	})

	t.Run("Create an Account", func(t *testing.T) {
		_, s, gw := setup()
		newAddress := flow.HexToAddress("192440c99cb17282")

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.FlowTransaction().Authorizers[0], serviceAddress)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		compareAddress := serviceAddress
		gw.GetAccount.Run(func(args mock.Arguments) {
			address := args.Get(0).(flow.Address)
			assert.Equal(t, address, compareAddress)
			compareAddress = newAddress
			gw.GetAccount.Return(
				tests.NewAccountWithAddress(address.String()), nil,
			)
		})

		gw.GetTransactionResult.Return(
			tests.NewAccountCreateResult(newAddress), nil,
		)

		account, err := s.Accounts.Create(
			serviceAcc,
			[]crypto.PublicKey{pubKey},
			[]int{1000},
			[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
			[]crypto.HashAlgorithm{crypto.SHA3_256},
			nil,
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, newAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.Equal(t, account.Address, newAddress)
		assert.NoError(t, err)
	})

	t.Run("Create an Account with Contract", func(t *testing.T) {
		_, s, gw := setup()
		newAddress := flow.HexToAddress("192440c99cb17281")

		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.FlowTransaction().Authorizers[0], serviceAddress)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "acct.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		gw.GetTransactionResult.Run(func(args mock.Arguments) {
			gw.GetTransactionResult.Return(tests.NewAccountCreateResult(newAddress), nil)
		})

		account, err := s.Accounts.Create(
			serviceAcc,
			[]crypto.PublicKey{pubKey},
			[]int{1000},
			[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
			[]crypto.HashAlgorithm{crypto.SHA3_256},
			[]string{"Hello:contractHello.cdc"},
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertCalled(t, tests.GetAccountFunc, newAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.Equal(t, account.Address, newAddress)
		assert.NoError(t, err)
	})

	t.Run("Contract Add for Account", func(t *testing.T) {
		_, s, gw := setup()
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.add"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		account, err := s.Accounts.AddContract(
			serviceAcc,
			tests.ContractHelloString.Filename,
			tests.ContractHelloString.Source,
			false,
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.NoError(t, err)
	})

	t.Run("Contract Update for Account", func(t *testing.T) {
		_, s, gw := setup()
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.update__experimental"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		account, err := s.Accounts.AddContract(
			serviceAcc,
			tests.ContractHelloString.Filename,
			tests.ContractHelloString.Source,
			true,
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.NoError(t, err)
	})

	t.Run("Contract Remove for Account", func(t *testing.T) {
		_, s, gw := setup()
		gw.SendSignedTransaction.Run(func(args mock.Arguments) {
			tx := args.Get(0).(*flowkit.Transaction)
			assert.Equal(t, tx.Signer().Address(), serviceAddress)
			assert.True(t, strings.Contains(string(tx.FlowTransaction().Script), "signer.contracts.remove"))

			gw.SendSignedTransaction.Return(tests.NewTransaction(), nil)
		})

		account, err := s.Accounts.RemoveContract(
			serviceAcc,
			tests.ContractHelloString.Filename,
		)

		gw.Mock.AssertCalled(t, tests.GetAccountFunc, serviceAddress)
		gw.Mock.AssertNumberOfCalls(t, tests.GetAccountFunc, 2)
		gw.Mock.AssertNumberOfCalls(t, tests.GetTransactionResultFunc, 1)
		gw.Mock.AssertNumberOfCalls(t, tests.SendSignedTransactionFunc, 1)
		assert.NotNil(t, account)
		assert.NoError(t, err)
	})

	t.Run("Staking Info for Account", func(t *testing.T) {
		_, s, gw := setup()

		count := 0
		gw.ExecuteScript.Run(func(args mock.Arguments) {
			count++
			assert.True(t, strings.Contains(string(args.Get(0).([]byte)), "import FlowIDTableStaking from 0x9eca2b38b18b5dfe"))
			gw.ExecuteScript.Return(cadence.NewArray([]cadence.Value{}), nil)
		})

		val1, val2, err := s.Accounts.StakingInfo(flow.HexToAddress("df9c30eb2252f1fa"))
		assert.NoError(t, err)
		assert.NotNil(t, val1)
		assert.NotNil(t, val2)
		assert.Equal(t, 2, count)
	})
	t.Run("Staking Info for Account fetches node total", func(t *testing.T) {
		_, s, gw := setup()

		count := 0
		gw.ExecuteScript.Run(func(args mock.Arguments) {
			assert.True(t, strings.Contains(string(args.Get(0).([]byte)), "import FlowIDTableStaking from 0x9eca2b38b18b5dfe"))
			if count < 2 {
				gw.ExecuteScript.Return(cadence.NewArray(
					[]cadence.Value{
						cadence.Struct{
							StructType: &cadence.StructType{
								Fields: []cadence.Field{
									{
										Identifier: "id",
									},
								},
							},
							Fields: []cadence.Value{
								cadence.String("8f4d09dae7918afbf62c48fa968a9e8b0891cee8442065fa47cc05f4bc9a8a91"),
							},
						},
					}), nil)
			} else {
				assert.True(t, strings.Contains(args.Get(1).([]cadence.Value)[0].String(), "8f4d09dae7918afbf62c48fa968a9e8b0891cee8442065fa47cc05f4bc9a8a91"))
				gw.ExecuteScript.Return(cadence.NewUFix64("1.0"))
			}
			count++
		})

		val1, val2, err := s.Accounts.StakingInfo(flow.HexToAddress("df9c30eb2252f1fa"))
		assert.NoError(t, err)
		assert.NotNil(t, val1)
		assert.NotNil(t, val2)
		assert.Equal(t, 3, count)
	})
}

func setupIntegration() (*flowkit.State, *Services) {
	readerWriter := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	acc, _ := state.EmulatorServiceAccount()
	gw := gateway.NewEmulatorGateway(acc)
	s := NewServices(gw, state, output.NewStdoutLogger(output.NoneLog))

	return state, s
}

func TestAccountsCreate_Integration(t *testing.T) {
	t.Parallel()

	type accountsIn struct {
		account  *flowkit.Account
		pubKeys  []crypto.PublicKey
		weights  []int
		sigAlgo  []crypto.SignatureAlgorithm
		hashAlgo []crypto.HashAlgorithm
		args     []string
	}

	type accountsOut struct {
		address string
		code    map[string][]byte
		balance uint64
		pubKeys []crypto.PublicKey
		weights []int
	}

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		accIn := []accountsIn{{
			account: srvAcc,
			sigAlgo: []crypto.SignatureAlgorithm{
				tests.SigAlgos()[0],
			},
			hashAlgo: []crypto.HashAlgorithm{
				tests.HashAlgos()[0],
			},
			args: nil,
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}, {
			account: srvAcc,
			args:    nil,
			sigAlgo: []crypto.SignatureAlgorithm{
				tests.SigAlgos()[0],
				tests.SigAlgos()[1],
			},
			hashAlgo: []crypto.HashAlgorithm{
				tests.HashAlgos()[0],
				tests.HashAlgos()[1],
			},
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
				tests.PubKeys()[1],
			},
			weights: []int{500, 500},
		}, {
			account: srvAcc,
			args: []string{
				fmt.Sprintf(
					"Simple:%s",
					tests.ContractSimple.Filename,
				),
			},
			sigAlgo: []crypto.SignatureAlgorithm{
				tests.SigAlgos()[0],
			},
			hashAlgo: []crypto.HashAlgorithm{
				tests.HashAlgos()[0],
			},
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}}

		accOut := []accountsOut{{
			address: "01cf0e2f2f715450",
			code:    map[string][]byte{},
			balance: uint64(100000),
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}, {
			address: "179b6b1cb6755e31",
			code:    map[string][]byte{},
			balance: uint64(100000),
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
				tests.PubKeys()[1],
			},
			weights: []int{500, 500},
		}, {
			address: "f3fcd2c1a78f5eee",
			code: map[string][]byte{
				tests.ContractSimple.Name: tests.ContractSimple.Source,
			},
			balance: uint64(100000),
			pubKeys: []crypto.PublicKey{
				tests.PubKeys()[0],
			},
			weights: []int{flow.AccountKeyWeightThreshold},
		}}

		for i, a := range accIn {
			acc, err := s.Accounts.Create(a.account, a.pubKeys, a.weights, a.sigAlgo, a.hashAlgo, a.args)
			c := accOut[i]

			assert.NoError(t, err)
			assert.NotNil(t, acc)
			assert.Equal(t, acc.Address.String(), c.address)
			assert.Equal(t, acc.Contracts, c.code)
			assert.Equal(t, acc.Balance, c.balance)
			assert.Len(t, acc.Keys, len(c.pubKeys))

			for x, k := range acc.Keys {
				assert.Equal(t, k.PublicKey, a.pubKeys[x])
				assert.Equal(t, k.Weight, c.weights[x])
				assert.Equal(t, k.SigAlgo, a.sigAlgo[x])
				assert.Equal(t, k.HashAlgo, a.hashAlgo[x])
			}

		}

	})

	t.Run("Create Invalid", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		errOut := []string{
			"open Invalid: file does not exist",
			"invalid account key: signing algorithm (UNKNOWN) is incompatible with hashing algorithm (SHA3_256)",
			"invalid account key: signing algorithm (UNKNOWN) is incompatible with hashing algorithm (UNKNOWN)",
			"number of keys and weights provided must match, number of provided keys: 2, number of provided key weights: 1",
			"number of keys and weights provided must match, number of provided keys: 1, number of provided key weights: 2",
		}

		accIn := []accountsIn{
			{
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.ECDSA_P256},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				args:     []string{"Invalid:Invalid"},
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.UnknownSignatureAlgorithm},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				args:     nil,
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.UnknownSignatureAlgorithm},
				hashAlgo: []crypto.HashAlgorithm{crypto.UnknownHashAlgorithm},
				args:     nil,
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.ECDSA_P256},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				args:     nil,
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
					tests.PubKeys()[1],
				},
				weights: []int{1000},
			}, {
				account:  srvAcc,
				sigAlgo:  []crypto.SignatureAlgorithm{crypto.ECDSA_P256},
				hashAlgo: []crypto.HashAlgorithm{crypto.SHA3_256},
				args:     nil,
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{1000, 1000},
			},
			/*{
			 	TODO(sideninja): uncomment this test case after https://github.com/onflow/flow-go-sdk/pull/199 is released
				account:  srvAcc,
				sigAlgo:  crypto.ECDSA_P256,
				hashAlgo: crypto.SHA3_256,
				args:     nil,
				pubKeys: []crypto.PublicKey{
					tests.PubKeys()[0],
				},
				weights: []int{-1},
			}*/
		}

		for i, a := range accIn {
			acc, err := s.Accounts.Create(a.account, a.pubKeys, a.weights, a.sigAlgo, a.hashAlgo, a.args)
			errMsg := errOut[i]

			assert.Nil(t, acc)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), errMsg)
		}
	})

}

func TestAccountsAddContract_Integration(t *testing.T) {
	t.Parallel()

	t.Run("Add Contract", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		acc, err := s.Accounts.AddContract(srvAcc, tests.ContractSimple.Name, tests.ContractSimple.Source, false)

		assert.NoError(t, err)
		assert.NotNil(t, acc)
		assert.Equal(t, acc.Contracts["Simple"], tests.ContractSimple.Source)

		acc, err = s.Accounts.AddContract(srvAcc, tests.ContractSimpleUpdated.Name, tests.ContractSimpleUpdated.Source, true)

		assert.NoError(t, err)
		assert.NotNil(t, acc)
		assert.Equal(t, acc.Contracts["Simple"], tests.ContractSimpleUpdated.Source)
	})

	t.Run("Add Contract Invalid", func(t *testing.T) {
		t.Parallel()

		state, s := setupIntegration()
		srvAcc, _ := state.EmulatorServiceAccount()

		// prepare existing contract
		_, err := s.Accounts.AddContract(srvAcc, tests.ContractSimple.Name, tests.ContractSimple.Source, false)
		assert.NoError(t, err)

		_, err = s.Accounts.AddContract(srvAcc, tests.ContractSimple.Name, tests.ContractSimple.Source, false)
		assert.True(t, strings.Contains(err.Error(), "cannot overwrite existing contract with name \"Simple\""))

		_, err = s.Accounts.AddContract(srvAcc, tests.ContractHelloString.Name, tests.ContractHelloString.Source, true)
		assert.True(t, strings.Contains(err.Error(), "cannot update non-existing contract with name \"Hello\""))
	})
}

func TestAccountsRemoveContract_Integration(t *testing.T) {
	t.Parallel()

	state, s := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	// prepare existing contract
	_, err := s.Accounts.AddContract(srvAcc, tests.ContractSimple.Name, tests.ContractSimple.Source, false)
	assert.NoError(t, err)

	t.Run("Remove Contract", func(t *testing.T) {
		t.Parallel()

		acc, err := s.Accounts.RemoveContract(srvAcc, tests.ContractSimple.Name)

		assert.NoError(t, err)
		assert.Equal(t, acc.Contracts[tests.ContractSimple.Name], []byte(nil))
	})
}

func TestAccountsGet_Integration(t *testing.T) {
	t.Parallel()

	state, s := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	t.Run("Get Account", func(t *testing.T) {
		t.Parallel()
		acc, err := s.Accounts.Get(srvAcc.Address())

		assert.NoError(t, err)
		assert.NotNil(t, acc)
		assert.Equal(t, acc.Address, srvAcc.Address())
	})

	t.Run("Get Account Invalid", func(t *testing.T) {
		t.Parallel()

		acc, err := s.Accounts.Get(flow.HexToAddress("0x1"))
		assert.Nil(t, acc)
		assert.Equal(t, err.Error(), "could not find account with address 0000000000000001")
	})
}

func TestAccountsStakingInfo_Integration(t *testing.T) {
	t.Parallel()
	state, s := setupIntegration()
	srvAcc, _ := state.EmulatorServiceAccount()

	t.Run("Get Staking Info", func(t *testing.T) {
		_, _, err := s.Accounts.StakingInfo(srvAcc.Address()) // unfortunately can't do integration test
		assert.Equal(t, err.Error(), "emulator chain not supported")
	})
}
