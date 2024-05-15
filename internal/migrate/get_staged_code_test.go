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
	"github.com/onflow/contract-updater/lib/go/templates"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_GetStagedCode(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Success", func(t *testing.T) {
		testContract := tests.ContractSimple

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
		account, err := state.EmulatorServiceAccount()
		assert.NoError(t, err)

		srv.Network.Return(config.Network{
			Name: "testnet",
		}, nil)

		srv.ExecuteScript.Run(func(args mock.Arguments) {
			script := args.Get(1).(flowkit.Script)

			actualContractAddressArg, actualContractNameArg := script.Args[0], script.Args[1]

			assert.Equal(t, templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress("testnet")), script.Code)

			assert.Equal(t, 2, len(script.Args))

			assert.EqualValues(t, testContract.Name, actualContractNameArg)

			contractAddr := cadence.NewAddress(account.Address)
			assert.Equal(t, contractAddr.String(), actualContractAddressArg.String())
		}).Return(cadence.NewOptional(cadence.String(testContract.Source)), nil)

		result, err := getStagedCode(
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
}
