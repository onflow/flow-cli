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

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_StageContract(t *testing.T) {
	testContract := tests.ContractSimple

	t.Run("Success", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)

		// Add contract to state
		state.Contracts().AddOrUpdate(
			config.Contract{
				Name:     testContract.Name,
				Location: testContract.Filename,
			},
		)

		// Add deployment to state
		state.Deployments().AddOrUpdate(
			config.Deployment{
				Network: "testnet",
				Account: "emulator-account",
				Contracts: []config.ContractDeployment{
					{
						Name: testContract.Name,
					},
				},
			},
		)

		srv.Network.Return(config.Network{
			Name: "testnet",
		}, nil)

		srv.Mock.On("ReplaceImportsInScript", mock.Anything, mock.Anything).Return(
			flowkit.Script{
				Code: testContract.Source,
			}, nil)

		srv.SendTransaction.Run(func(args mock.Arguments) {
			script := args.Get(2).(flowkit.Script)

			actualContractNameArg, actualContractCodeArg := script.Args[0], script.Args[1]

			contractName, _ := cadence.NewString(testContract.Name)
			contractBody, _ := cadence.NewString(string(testContract.Source))
			assert.Equal(t, contractName, actualContractNameArg)
			assert.Equal(t, contractBody, actualContractCodeArg)
		}).Return(flow.NewTransaction(), &flow.TransactionResult{
			Status:      flow.TransactionStatusSealed,
			Error:       nil,
			BlockHeight: 1,
		}, nil)

		result, err := stageContract(
			[]string{testContract.Name},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("missing contract", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)
		result, err := stageContract(
			[]string{testContract.Name},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
