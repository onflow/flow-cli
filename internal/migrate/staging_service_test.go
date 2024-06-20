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

package migrate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/contract-updater/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	flowkitMocks "github.com/onflow/flowkit/v2/mocks"
	"github.com/onflow/flowkit/v2/project"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/internal/util"
)

type mockDeployment struct {
	name string
	code string
}

type mockAccount struct {
	name        string
	address     string
	deployments []mockDeployment
}

func addAccountsToState(
	t *testing.T,
	state *flowkit.State,
	accts []mockAccount,
) {
	for _, account := range accts {
		key, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, make([]byte, 32))
		require.NoError(t, err)

		state.Accounts().AddOrUpdate(
			&accounts.Account{
				Name:    account.name,
				Address: flow.HexToAddress(account.address),
				Key: accounts.NewHexKeyFromPrivateKey(
					0,
					crypto.SHA3_256,
					key,
				),
			},
		)

		contractDeployments := make([]config.ContractDeployment, 0)
		for _, deployment := range account.deployments {
			fname := account.address + "/" + deployment.name + ".cdc"
			require.NoError(t, state.ReaderWriter().WriteFile(fname, []byte(deployment.code), 0644))

			state.Contracts().AddOrUpdate(
				config.Contract{
					Name:     deployment.name,
					Location: fname,
				},
			)

			contractDeployments = append(
				contractDeployments,
				config.ContractDeployment{
					Name: deployment.name,
				},
			)
		}

		state.Deployments().AddOrUpdate(
			config.Deployment{
				Network:   "testnet",
				Account:   account.name,
				Contracts: contractDeployments,
			},
		)
	}
}

func Test_StagingService(t *testing.T) {
	setupMocks := func(
		accts []mockAccount,
		mocksStagedContracts map[common.AddressLocation][]byte,
		txResult *flow.TransactionResult,
	) (*flowkitMocks.Services, *flowkit.State, []*project.Contract) {
		srv := flowkitMocks.NewServices(t)
		rw, _ := tests.ReaderWriter()
		state, err := flowkit.Init(rw)
		require.NoError(t, err)

		addAccountsToState(t, state, accts)

		srv.On("Network", mock.Anything).Return(config.Network{
			Name: "testnet",
		}, nil).Maybe()

		srv.On("ReplaceImportsInScript", mock.Anything, mock.Anything).Return(func(_ context.Context, script flowkit.Script) (flowkit.Script, error) {
			return script, nil
		}).Maybe()

		deploymentContracts, err := state.DeploymentContractsByNetwork(config.TestnetNetwork)
		require.NoError(t, err)

		srv.On("SendTransaction", mock.Anything, mock.Anything, mock.MatchedBy(func(script flowkit.Script) bool {
			expectedScript := templates.GenerateStageContractScript(MigrationContractStagingAddress("testnet"))
			if string(script.Code) != string(expectedScript) {
				return false
			}

			if len(script.Args) != 2 {
				return false
			}

			_, ok := script.Args[0].(cadence.String)
			if !ok {
				return false
			}

			_, ok = script.Args[1].(cadence.String)
			return ok
		}), mock.Anything).Return(tests.NewTransaction(), txResult, nil).Maybe()

		// Mock staged contracts on network
		for location, code := range mocksStagedContracts {
			srv.On(
				"ExecuteScript",
				mock.Anything,
				mock.MatchedBy(func(script flowkit.Script) bool {
					if string(script.Code) != string(templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress("testnet"))) {
						return false
					}

					if len(script.Args) != 2 {
						return false
					}

					callContractAddress, callContractName := script.Args[0], script.Args[1]
					if callContractName != cadence.String(location.Name) {
						return false
					}
					if callContractAddress != cadence.Address(location.Address) {
						return false
					}

					return true
				}),
				mock.Anything,
			).Return(cadence.NewOptional(cadence.String(code)), nil).Maybe()
		}

		// Default all staged contracts to nil
		srv.On(
			"ExecuteScript",
			mock.Anything,
			mock.MatchedBy(func(script flowkit.Script) bool {
				if string(script.Code) != string(templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress("testnet"))) {
					return false
				}

				if len(script.Args) != 2 {
					return false
				}

				return true
			}),
			mock.Anything,
		).Return(cadence.NewOptional(nil), nil).Maybe()

		return srv, state, deploymentContracts
	}

	t.Run("stages valid contracts", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
					{
						name: "Bar",
						code: "access(all) contract Bar {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, nil, tests.NewTransactionResult(nil))

		v := newMockStagingValidator(t)
		v.On("Validate", mock.MatchedBy(func(stagedContracts []stagedContractUpdate) bool {
			return reflect.DeepEqual(stagedContracts, []stagedContractUpdate{
				{
					DeployLocation: simpleAddressLocation("0x01.Foo"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Foo.cdc")),
					Code:           []byte("access(all) contract Foo {}"),
				},
				{
					DeployLocation: simpleAddressLocation("0x01.Bar"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Bar.cdc")),
					Code:           []byte("access(all) contract Bar {}"),
				},
			})
		})).Return(nil).Once()

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			v,
			func(sve *stagingValidatorError) bool {
				return false
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 2, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].Err)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].WasValidated, true)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].TxId)

		require.Contains(t, results, simpleAddressLocation("0x01.Bar"))
		require.Nil(t, results[simpleAddressLocation("0x01.Bar")].Err)
		require.Equal(t, results[simpleAddressLocation("0x01.Bar")].WasValidated, true)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Bar")].TxId)

		srv.AssertNumberOfCalls(t, "SendTransaction", 2)
	})

	t.Run("stages unvalidated contracts if chosen", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, nil, tests.NewTransactionResult(nil))

		v := newMockStagingValidator(t)
		v.On("Validate", mock.MatchedBy(func(stagedContracts []stagedContractUpdate) bool {
			return reflect.DeepEqual(stagedContracts, []stagedContractUpdate{
				{
					DeployLocation: simpleAddressLocation("0x01.Foo"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Foo.cdc")),
					Code:           []byte("access(all) contract Foo {}"),
				},
			})
		})).Return(&stagingValidatorError{
			errors: map[common.AddressLocation]error{
				simpleAddressLocation("0x01.Foo"): &missingDependenciesError{
					MissingContracts: []common.AddressLocation{
						simpleAddressLocation("0x02.Bar"),
					},
				},
			},
		}).Once()

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			v,
			func(sve *stagingValidatorError) bool {
				require.NotNil(t, sve)
				return true
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 1, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].Err)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].WasValidated, false)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].TxId)

		srv.AssertNumberOfCalls(t, "SendTransaction", 1)
	})

	t.Run("skips validation if no validator", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, nil, tests.NewTransactionResult(nil))

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			nil,
			func(sve *stagingValidatorError) bool {
				require.NotNil(t, sve)
				return true
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 1, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].Err)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].WasValidated, false)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].TxId)

		srv.AssertNumberOfCalls(t, "SendTransaction", 1)
	})

	t.Run("returns missing dependency error if staging not chosen", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, nil, tests.NewTransactionResult(nil))

		v := newMockStagingValidator(t)
		v.On("Validate", mock.MatchedBy(func(stagedContracts []stagedContractUpdate) bool {
			return reflect.DeepEqual(stagedContracts, []stagedContractUpdate{
				{
					DeployLocation: simpleAddressLocation("0x01.Foo"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Foo.cdc")),
					Code:           []byte("access(all) contract Foo {}"),
				},
			})
		})).Return(&stagingValidatorError{
			errors: map[common.AddressLocation]error{
				simpleAddressLocation("0x01.Foo"): &missingDependenciesError{
					MissingContracts: []common.AddressLocation{
						simpleAddressLocation("0x02.Bar"),
					},
				},
			},
		}).Once()

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			v,
			func(sve *stagingValidatorError) bool {
				require.NotNil(t, sve)
				return false
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 1, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))

		var mde *missingDependenciesError
		require.ErrorAs(t, results[simpleAddressLocation("0x01.Foo")].Err, &mde)
		require.NotNil(t, results[simpleAddressLocation("0x01.Foo")].Err)
		require.Equal(t, []common.AddressLocation{simpleAddressLocation("0x02.Bar")}, mde.MissingContracts)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].WasValidated, true)

		srv.AssertNumberOfCalls(t, "SendTransaction", 0)
	})

	t.Run("reports and does not stage invalid contracts", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
					{
						name: "Bar",
						code: "access(all) contract Bar {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, nil, tests.NewTransactionResult(nil))

		v := newMockStagingValidator(t)
		v.On("Validate", mock.MatchedBy(func(stagedContracts []stagedContractUpdate) bool {
			return reflect.DeepEqual(stagedContracts, []stagedContractUpdate{
				{
					DeployLocation: simpleAddressLocation("0x01.Foo"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Foo.cdc")),
					Code:           []byte("access(all) contract Foo {}"),
				},
				{
					DeployLocation: simpleAddressLocation("0x01.Bar"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Bar.cdc")),
					Code:           []byte("access(all) contract Bar {}"),
				},
			})
		})).Return(&stagingValidatorError{
			errors: map[common.AddressLocation]error{
				simpleAddressLocation("0x01.Foo"): errors.New("FooError"),
			},
		}).Once()

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			v,
			func(sve *stagingValidatorError) bool {
				return false
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 2, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))
		require.ErrorContains(t, results[simpleAddressLocation("0x01.Foo")].Err, "FooError")
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].WasValidated, true)

		require.Contains(t, results, simpleAddressLocation("0x01.Bar"))
		require.Nil(t, results[simpleAddressLocation("0x01.Bar")].Err)
		require.Equal(t, results[simpleAddressLocation("0x01.Bar")].WasValidated, true)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Bar")].TxId)

		srv.AssertNumberOfCalls(t, "SendTransaction", 1)
	})

	t.Run("skips staging contracts without changes", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, map[common.AddressLocation][]byte{
			simpleAddressLocation("0x01.Foo"): []byte("access(all) contract Foo {}"),
		}, tests.NewTransactionResult(nil))

		v := newMockStagingValidator(t)
		v.On("Validate", mock.MatchedBy(func(stagedContracts []stagedContractUpdate) bool {
			return reflect.DeepEqual(stagedContracts, []stagedContractUpdate{
				{
					DeployLocation: simpleAddressLocation("0x01.Foo"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Foo.cdc")),
					Code:           []byte("access(all) contract Foo {}"),
				},
			})
		})).Return(nil).Once()

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			v,
			func(sve *stagingValidatorError) bool {
				return false
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 1, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].Err)
		require.Equal(t, true, results[simpleAddressLocation("0x01.Foo")].WasValidated)
		require.Equal(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].TxId)

		srv.AssertNumberOfCalls(t, "SendTransaction", 0)
	})

	t.Run("handles error transaction result", func(t *testing.T) {
		mockAccount := []mockAccount{
			{
				name:    "some-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: "access(all) contract Foo {}",
					},
				},
			},
		}
		srv, state, deploymentContracts := setupMocks(mockAccount, nil, &flow.TransactionResult{
			Status:        flow.TransactionStatusSealed,
			Error:         fmt.Errorf("i am a transaction error"),
			TransactionID: flow.Identifier{0x99},
		})

		v := newMockStagingValidator(t)
		v.On("Validate", mock.MatchedBy(func(stagedContracts []stagedContractUpdate) bool {
			return reflect.DeepEqual(stagedContracts, []stagedContractUpdate{
				{
					DeployLocation: simpleAddressLocation("0x01.Foo"),
					SourceLocation: common.StringLocation(filepath.FromSlash("0x01/Foo.cdc")),
					Code:           []byte("access(all) contract Foo {}"),
				},
			})
		})).Return(nil).Once()

		s := newStagingService(
			srv,
			state,
			util.NoLogger,
			v,
			func(sve *stagingValidatorError) bool {
				return false
			},
		)

		results, err := s.StageContracts(
			context.Background(),
			deploymentContracts,
		)

		require.NoError(t, err)
		require.NotNil(t, results)

		require.Equal(t, 1, len(results))
		require.Contains(t, results, simpleAddressLocation("0x01.Foo"))
		require.ErrorContains(t, results[simpleAddressLocation("0x01.Foo")].Err, "i am a transaction error")
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].WasValidated, true)
		require.Equal(t, flow.Identifier{0x99}, results[simpleAddressLocation("0x01.Foo")].TxId)

		srv.AssertNumberOfCalls(t, "SendTransaction", 1)
	})
}
