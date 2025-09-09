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

package dependencymanager

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func TestListDependencies(t *testing.T) {
	t.Run("Empty dependencies", func(t *testing.T) {
		logger := output.NewStdoutLogger(output.NoneLog)
		_, state, _ := util.TestMocks(t)

		result, err := list([]string{}, command.GlobalFlags{}, logger, nil, state)

		assert.NoError(t, err)
		listResult, ok := result.(*ListResult)
		assert.True(t, ok)
		assert.Equal(t, 0, len(listResult.Dependencies))
	})

	t.Run("With dependencies", func(t *testing.T) {
		logger := output.NewStdoutLogger(output.NoneLog)
		_, state, _ := util.TestMocks(t)

		serviceAcc, _ := state.EmulatorServiceAccount()
		dep := config.Dependency{
			Name: "TestContract",
			Source: config.Source{
				NetworkName:  "emulator",
				Address:      serviceAcc.Address,
				ContractName: "TestContract",
			},
		}

		state.Dependencies().AddOrUpdate(dep)

		result, err := list([]string{}, command.GlobalFlags{}, logger, nil, state)

		assert.NoError(t, err)
		listResult, ok := result.(*ListResult)
		assert.True(t, ok)
		assert.Equal(t, 1, len(listResult.Dependencies))

		depInfo := listResult.Dependencies[0]
		assert.Equal(t, "TestContract", depInfo.Name)
		assert.Equal(t, "emulator", depInfo.NetworkName)
		assert.Equal(t, serviceAcc.Address.String(), depInfo.Address)
		assert.Equal(t, "TestContract", depInfo.Contract)
	})
}

func TestListResult_JSON(t *testing.T) {
	t.Run("JSON output", func(t *testing.T) {
		result := &ListResult{
			Dependencies: []DependencyInfo{
				{
					Name:        "TestContract",
					NetworkName: "emulator",
					Address:     "0x01cf0e2f2f715450",
					Contract:    "TestContract",
				},
			},
		}

		jsonOutput := result.JSON()
		assert.Equal(t, result, jsonOutput)
	})
}
