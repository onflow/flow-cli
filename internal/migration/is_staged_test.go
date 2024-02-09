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

package migration

import (
	"testing"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func Test_IsStaged(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	testContract := tests.ContractSimple

	t.Run("Success", func(t *testing.T) {

		srv.ExecuteScript.Run(func(args mock.Arguments) {
			script := args.Get(1).(flowkit.Script)

			actualContractAddressArg, actualContractNameArg := script.Args[0], script.Args[1]

			contractName, _ := cadence.NewString(testContract.Name)
			contractAddr := cadence.NewAddress(flowsdk.HexToAddress("0xSomeAddress"))
			assert.Equal(t, contractName, actualContractNameArg)
			assert.Equal(t, contractAddr, actualContractAddressArg)
		}).Return(cadence.NewMeteredBool(nil, true), nil)

		result, err := isStaged(
			[]string{testContract.Name, "0xSomeAddress"},
			command.GlobalFlags{
				Network: "testnet",
			},
			util.NoLogger,
			srv.Mock,
			state,
		)
		assert.NoError(t, err)
		// TODO: fix this
		assert.NotNil(t, result)
	})
}
