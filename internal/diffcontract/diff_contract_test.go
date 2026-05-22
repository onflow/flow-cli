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

package diffcontract

import (
	"context"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/mocks"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

const testContractCode = `access(all) contract TestContract {
    access(all) fun hello(): String {
        return "Hello"
    }
}
`

const testContractCodeModified = `access(all) contract TestContract {
    access(all) fun hello(): String {
        return "Hello, World!"
    }
}
`

// passthrough is a mock return function that returns the script unchanged
func passthrough(_ context.Context, s flowkit.Script) (flowkit.Script, error) {
	return s, nil
}

func setupMocks(t *testing.T, deployedCode []byte, localCode []byte) (*mocks.MockServices, *flowkit.State) {
	srv, state, rw := util.TestMocks(t)

	// Write contract file to mock filesystem
	err := rw.WriteFile("TestContract.cdc", localCode, 0644)
	require.NoError(t, err)

	// Mock ReplaceImportsInScript to return code as-is (no imports to resolve)
	srv.Mock.On(
		"ReplaceImportsInScript",
		mock.Anything,
		mock.AnythingOfType("flowkit.Script"),
	).Return(passthrough)

	// Mock GetAccount to return account with deployed contract
	account := &flow.Account{
		Address:   flow.HexToAddress("f8d6e0586b0a20c7"),
		Contracts: map[string][]byte{"TestContract": deployedCode},
	}
	srv.GetAccount.Run(func(args mock.Arguments) {
		srv.GetAccount.Return(account, nil)
	})

	return srv, state
}

func Test_DiffContract(t *testing.T) {
	t.Run("Identical contracts", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCode))
		diffFlags.Quiet = false

		result, err := diffContract(
			[]string{"TestContract.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		r := result.(*diffContractResult)
		assert.True(t, r.identical)
		assert.Equal(t, 0, r.ExitCode())
		assert.Contains(t, r.String(), "up to date")
		assert.Equal(t, "identical", r.Oneliner())
	})

	t.Run("Different contracts", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCodeModified))
		diffFlags.Quiet = false

		result, err := diffContract(
			[]string{"TestContract.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		r := result.(*diffContractResult)
		assert.False(t, r.identical)
		assert.Equal(t, 1, r.ExitCode())
		assert.NotEmpty(t, r.String())
		assert.Contains(t, r.String(), "---")
		assert.Contains(t, r.String(), "+++")
		assert.Contains(t, r.String(), "@@")
		assert.Equal(t, "different", r.Oneliner())
	})

	t.Run("Quiet mode identical", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCode))
		diffFlags.Quiet = true

		result, err := diffContract(
			[]string{"TestContract.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		r := result.(*diffContractResult)
		assert.Equal(t, 0, r.ExitCode())
		assert.Equal(t, "", r.String())
	})

	t.Run("Quiet mode different", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCodeModified))
		diffFlags.Quiet = true

		result, err := diffContract(
			[]string{"TestContract.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		r := result.(*diffContractResult)
		assert.Equal(t, 1, r.ExitCode())
		assert.Equal(t, "", r.String())
	})

	t.Run("Contract not found on account", func(t *testing.T) {
		srv, state, rw := util.TestMocks(t)

		err := rw.WriteFile("TestContract.cdc", []byte(testContractCode), 0644)
		require.NoError(t, err)

		srv.Mock.On(
			"ReplaceImportsInScript",
			mock.Anything,
			mock.AnythingOfType("flowkit.Script"),
		).Return(passthrough)

		// Account with no contracts
		account := &flow.Account{
			Address:   flow.HexToAddress("f8d6e0586b0a20c7"),
			Contracts: map[string][]byte{},
		}
		srv.GetAccount.Run(func(args mock.Arguments) {
			srv.GetAccount.Return(account, nil)
		})

		result, err := diffContract(
			[]string{"TestContract.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.EqualError(t, err, "contract 'TestContract' not found on account f8d6e0586b0a20c7")
	})

	t.Run("Non-existing file", func(t *testing.T) {
		srv, state, _ := util.TestMocks(t)

		result, err := diffContract(
			[]string{"non-existing.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "error loading contract file")
	})

	t.Run("Resolve address from flow.json", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCode))
		diffFlags.Quiet = false

		// Add deployment config: emulator-account deploys TestContract on emulator
		state.Deployments().AddOrUpdate(config.Deployment{
			Network: "emulator",
			Account: "emulator-account",
			Contracts: []config.ContractDeployment{
				{Name: "TestContract"},
			},
		})

		// No address argument — should resolve from flow.json
		result, err := diffContract(
			[]string{"TestContract.cdc"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		r := result.(*diffContractResult)
		assert.True(t, r.identical)
		assert.Equal(t, 0, r.ExitCode())
	})

	t.Run("Resolve address from contract alias", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCode))
		diffFlags.Quiet = false

		// Add contract with alias instead of deployment
		state.Contracts().AddOrUpdate(config.Contract{
			Name:     "TestContract",
			Location: "TestContract.cdc",
			Aliases:  config.Aliases{{Network: "emulator", Address: flow.HexToAddress("f8d6e0586b0a20c7")}},
		})

		// No address argument — should resolve from alias
		result, err := diffContract(
			[]string{"TestContract.cdc"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		r := result.(*diffContractResult)
		assert.True(t, r.identical)
		assert.Equal(t, 0, r.ExitCode())
	})

	t.Run("No address and not in flow.json", func(t *testing.T) {
		srv, state, rw := util.TestMocks(t)

		err := rw.WriteFile("TestContract.cdc", []byte(testContractCode), 0644)
		require.NoError(t, err)

		srv.Mock.On(
			"ReplaceImportsInScript",
			mock.Anything,
			mock.AnythingOfType("flowkit.Script"),
		).Return(passthrough)

		// No deployment config, no address argument
		result, err := diffContract(
			[]string{"TestContract.cdc"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found in deployments or aliases")
	})

	t.Run("JSON output", func(t *testing.T) {
		srv, state := setupMocks(t, []byte(testContractCode), []byte(testContractCodeModified))
		diffFlags.Quiet = false

		result, err := diffContract(
			[]string{"TestContract.cdc", "f8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			srv.Mock,
			state,
		)

		require.NoError(t, err)

		jsonResult := result.JSON().(map[string]any)
		assert.Equal(t, "TestContract", jsonResult["contract"])
		assert.Equal(t, false, jsonResult["identical"])
		assert.NotEmpty(t, jsonResult["diff"])
	})
}
