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

	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/contract-updater/lib/go/templates"
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
		}`
		newContract := `
		access(all) contract Test {
			access(all) fun test() {}
		}`
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

		validator := newStagingValidator(srv.Mock, state)
		err := validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		require.NoError(t, err)
	})

	t.Run("contract update with update error", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		access(all) contract Test {
			access(all) let x: Int
			access(all) fun test() {}

			init() {
				self.x = 1
			}
		}`
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

		validator := newStagingValidator(srv.Mock, state)
		err := validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		var updateErr *stdlib.ContractUpdateError
		require.ErrorAs(t, err, &updateErr)
	})

	t.Run("contract update with checker error", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			let x: Int
			init() {
				self.x = 1
			}
		}`
		newContract := `
		access(all) contract Test {
			access(all) let x: Int
			init() {
				self.x = "bad type :("
			}
		}`
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

		validator := newStagingValidator(srv.Mock, state)
		err := validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		var checkerErr *sema.CheckerError
		require.ErrorAs(t, err, &checkerErr)
	})

	t.Run("valid contract update with dependencies", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import ImpContract from 0x02
		access(all) contract Test {
			access(all) fun test() {}
		}`
		impContract := `
		access(all) contract ImpContract {
			access(all) let x: Int
			init() {
				self.x = 1
			}
		}`
		mockScriptResultString, err := cadence.NewString(impContract)
		require.NoError(t, err)

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

			contractName, _ := cadence.NewString("ImpContract")
			contractAddr := cadence.NewAddress(flow.HexToAddress("02"))
			assert.Equal(t, contractName, actualContractNameArg)
			assert.Equal(t, contractAddr, actualContractAddressArg)
		}).Return(cadence.NewOptional(mockScriptResultString), nil)

		// validate
		validator := newStagingValidator(srv.Mock, state)
		err = validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		require.NoError(t, err)
	})

	t.Run("contract update missing dependency", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		impLocation := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}, "ImpContract")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import ImpContract from 0x02
		access(all) contract Test {
			access(all) fun test() {}
		}`
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
		srv.ExecuteScript.Return(cadence.NewOptional(nil), nil)

		validator := newStagingValidator(srv.Mock, state)
		err := validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		var missingDepsErr *missingDependenciesError
		require.ErrorAs(t, err, &missingDepsErr)
		require.Equal(t, 1, len(missingDepsErr.MissingContracts))
		require.Equal(t, impLocation, missingDepsErr.MissingContracts[0])
	})

	t.Run("valid contract update with system contract imports", func(t *testing.T) {
		location := common.NewAddressLocation(nil, common.Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "Test")
		sourceCodeLocation := common.StringLocation("./Test.cdc")
		oldContract := `
		import FlowToken from 0x7e60df042a9c0868
		pub contract Test {
			pub fun test() {}
		}`
		newContract := `
		import FlowToken from 0x7e60df042a9c0868
		import Burner from 0x9a0766d93b6608b7
		access(all) contract Test {
			access(all) fun test() {}
		}`
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

		validator := newStagingValidator(srv.Mock, state)
		err := validator.ValidateContractUpdate(location, sourceCodeLocation, []byte(newContract))
		require.NoError(t, err)
	})
}
