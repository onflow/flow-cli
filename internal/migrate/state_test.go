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

package migrate

import (
	"testing"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/cmd/util/ledger/migrations"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/tests"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/internal/util"
)

func Test_MigrateState(t *testing.T) {
	_, state, _ := util.TestMocks(t)

	testContractAliased := tests.ContractSimple
	testContractDeployed := tests.ContractA

	t.Run("resolves staged contracts by name", func(t *testing.T) {
		// Add an aliased contract to state
		state.Contracts().AddOrUpdate(
			config.Contract{
				Name:     testContractAliased.Name,
				Location: testContractAliased.Filename,
				Aliases: config.Aliases{
					{
						Network: "emulator",
						Address: flow.HexToAddress("0x1"),
					},
				},
			},
		)

		state.Contracts().AddOrUpdate(
			config.Contract{
				Name:     testContractDeployed.Name,
				Location: testContractDeployed.Filename,
			},
		)

		// Add deployment to state
		state.Deployments().AddOrUpdate(
			config.Deployment{
				Network: "emulator",
				Account: "emulator-account",
				Contracts: []config.ContractDeployment{
					{
						Name: testContractDeployed.Name,
					},
				},
			},
		)

		account, err := state.EmulatorServiceAccount()
		assert.NoError(t, err)

		contracts, err := resolveStagedContracts(
			state,
			[]string{testContractAliased.Name, testContractDeployed.Name},
		)
		assert.NoError(t, err)

		assert.Equal(t, []migrations.StagedContract{
			{
				Contract: migrations.Contract{
					Name: testContractAliased.Name,
					Code: testContractAliased.Source,
				},
				Address: common.Address(flow.HexToAddress("0x1")),
			},
			{
				Contract: migrations.Contract{
					Name: testContractDeployed.Name,
					Code: testContractDeployed.Source,
				},
				Address: common.Address(account.Address),
			},
		}, contracts)
	})
}
