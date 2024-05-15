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
	"testing"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	flowkitMocks "github.com/onflow/flowkit/v2/mocks"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_StageContract(t *testing.T) {
	setupMocks := func(
		accts []mockAccount,
	) (*mockStagingService, flowkit.Services, *flowkit.State) {
		ss := newMockStagingService(t)
		srv := flowkitMocks.NewServices(t)
		rw, _ := tests.ReaderWriter()
		state, _ := flowkit.Init(rw)

		addAccountsToState(t, state, accts)

		srv.On("Network").Return(config.Network{
			Name: "testnet",
		}, nil)

		return ss, srv, state
	}

	t.Run("all contracts filter", func(t *testing.T) {
		ss, srv, state := setupMocks([]mockAccount{
			{
				name:    "my-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: `FooCode`,
					},
				},
			},
		})

		mockResult := make(map[common.AddressLocation]stagingResult)
		mockResult[common.NewAddressLocation(nil, common.Address{0x01}, "Foo")] = stagingResult{
			Err: nil,
		}

		ss.On("StageContracts", mock.Anything, mock.Anything).Return(mockResult, nil)

		result, err := stageAll(
			ss,
			state,
			srv,
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, mockResult, result.Results)
	})

	t.Run("contract name filter", func(t *testing.T) {
		ss, srv, state := setupMocks([]mockAccount{
			{
				name:    "my-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: `FooCode`,
					},
					{
						name: "Bar",
						code: `BarCode`,
					},
				},
			},
		})

		mockResult := make(map[common.AddressLocation]stagingResult)
		mockResult[common.NewAddressLocation(nil, common.Address{0x01}, "Foo")] = stagingResult{
			Err: nil,
		}

		ss.On("StageContracts", mock.Anything, mock.Anything).Return(mockResult, nil).Once()

		result, err := stageByContractNames(
			ss,
			state,
			srv,
			[]string{"Foo"},
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, mockResult, result.Results)
	})

	t.Run("contract name filter", func(t *testing.T) {
		ss, srv, state := setupMocks([]mockAccount{
			{
				name:    "my-account",
				address: "0x01",
				deployments: []mockDeployment{
					{
						name: "Foo",
						code: `FooCode`,
					},
				},
			},
			{
				name:    "other-account",
				address: "0x02",
				deployments: []mockDeployment{
					{
						name: "Bar",
						code: `BarCode`,
					},
				},
			},
		})

		mockResult := make(map[common.AddressLocation]stagingResult)
		mockResult[common.NewAddressLocation(nil, common.Address{0x01}, "Foo")] = stagingResult{
			Err: nil,
		}
		ss.On("StageContracts", mock.Anything, mock.Anything).Return(mockResult, nil).Once()

		result, err := stageByAccountNames(
			ss,
			state,
			srv,
			[]string{"my-account"},
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, mockResult, result.Results)
	})

	t.Run("contract name not found", func(t *testing.T) {
		ss, srv, state := setupMocks(nil)

		result, err := stageByContractNames(
			ss,
			state,
			srv,
			[]string{"my-contract"},
		)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("account not found", func(t *testing.T) {
		ss, srv, state := setupMocks(nil)

		result, err := stageByAccountNames(
			ss,
			state,
			srv,
			[]string{"my-account"},
		)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
