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
	) (flowkit.Services, *flowkit.State, []*project.Contract) {
		srv := flowkitMocks.NewServices(t)
		rw, _ := tests.ReaderWriter()
		state, err := flowkit.Init(rw, crypto.ECDSA_P256, crypto.SHA3_256)
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
		}), mock.Anything).Return(tests.NewTransaction(), nil, nil).Maybe()

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
		srv, state, deploymentContracts := setupMocks(mockAccount)

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
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].err)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].wasValidated, true)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].txId)

		require.Contains(t, results, simpleAddressLocation("0x01.Bar"))
		require.Nil(t, results[simpleAddressLocation("0x01.Bar")].err)
		require.Equal(t, results[simpleAddressLocation("0x01.Bar")].wasValidated, true)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Bar")].txId)
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
		srv, state, deploymentContracts := setupMocks(mockAccount)

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
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].err)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].wasValidated, false)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].txId)
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
		srv, state, deploymentContracts := setupMocks(mockAccount)

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
		require.Nil(t, results[simpleAddressLocation("0x01.Foo")].err)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].wasValidated, false)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Foo")].txId)
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
		srv, state, deploymentContracts := setupMocks(mockAccount)

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
		require.ErrorAs(t, results[simpleAddressLocation("0x01.Foo")].err, &mde)
		require.NotNil(t, results[simpleAddressLocation("0x01.Foo")].err)
		require.Equal(t, []common.AddressLocation{simpleAddressLocation("0x02.Bar")}, mde.MissingContracts)
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].wasValidated, true)
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
		srv, state, deploymentContracts := setupMocks(mockAccount)

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
		require.ErrorContains(t, results[simpleAddressLocation("0x01.Foo")].err, "FooError")
		require.Equal(t, results[simpleAddressLocation("0x01.Foo")].wasValidated, true)

		require.Contains(t, results, simpleAddressLocation("0x01.Bar"))
		require.Nil(t, results[simpleAddressLocation("0x01.Bar")].err)
		require.Equal(t, results[simpleAddressLocation("0x01.Bar")].wasValidated, true)
		require.IsType(t, flow.Identifier{}, results[simpleAddressLocation("0x01.Bar")].txId)
	})
}
