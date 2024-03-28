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
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/contract-updater/lib/go/templates"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_StagingValidator(t *testing.T) {
	srv, state, rw := util.TestMocks(t)
	t.Run("valid contract update with no dependencies", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
			pub contract Test {
				pub fun test() {}
			}
		`
		newContract := `
			access(all) contract Test {
				access(all) fun test() {}
			}
		`
		mockAccount := &flow.Account{
			Address: flow.HexToAddress("01"),
			Balance: 1000,
			Keys:    nil,
			Contracts: map[string][]byte{
				"Test": []byte(oldContract),
			},
		}

		// setup mocks
		require.NoError(t, rw.WriteFile(sourceCodeLocation.String(), []byte(newContract), 0o644))
		srv.GetAccount.Run(func(args mock.Arguments) {
			require.Equal(t, flow.HexToAddress("01"), args.Get(1).(flow.Address))
		}).Return(mockAccount, nil)
		srv.Network.Return(config.Network{
			Name: "testnet",
		}, nil)
		srv.ExecuteScript.Run(func(args mock.Arguments) {
			script := args.Get(1).(flowkit.Script)

			assert.Equal(t, templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress("testnet")), script.Code)

			assert.Equal(t, 2, len(script.Args))
			actualContractAddressArg, actualContractNameArg := script.Args[0], script.Args[1]

			contractName, _ := cadence.NewString("Test")
			contractAddr := cadence.NewAddress(flow.HexToAddress("01"))
			assert.Equal(t, contractName, actualContractNameArg)
			assert.Equal(t, contractAddr, actualContractAddressArg)
		}).Return(cadence.NewBool(true), nil)

		validator := newStagingValidator(srv.Mock, state)
		err := validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		require.NoError(t, err)
	})
}
