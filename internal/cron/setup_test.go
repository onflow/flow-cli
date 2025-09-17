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

package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

func TestSetupCommand(t *testing.T) {
	t.Run("Should require flow configuration", func(t *testing.T) {
		result, err := setupRun(
			[]string{"0x123456789abcdef"},
			command.GlobalFlags{Network: "emulator"},
			output.NewStdoutLogger(output.NoneLog),
			nil,
			nil, // no state
		)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "flow configuration is required")
	})

	t.Run("Should require account argument", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		result, err := setupRun(
			[]string{},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			nil,
			state,
		)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "account is required as an argument")
	})

	t.Run("Should succeed with valid address", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		result, err := setupRun(
			[]string{"0xf8d6e0586b0a20c7"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			nil,
			state,
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		setupRes, ok := result.(*setupResult)
		require.True(t, ok)
		assert.True(t, setupRes.success)
		assert.Contains(t, setupRes.message, "created successfully")
	})

	t.Run("Should succeed with valid account name", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		result, err := setupRun(
			[]string{"emulator-account"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			nil,
			state,
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		setupRes, ok := result.(*setupResult)
		require.True(t, ok)
		assert.True(t, setupRes.success)
		assert.Contains(t, setupRes.message, "created successfully")
	})

	t.Run("Should error with invalid account name", func(t *testing.T) {
		_, state, _ := util.TestMocks(t)

		result, err := setupRun(
			[]string{"non-existent-account"},
			command.GlobalFlags{Network: "emulator"},
			util.NoLogger,
			nil,
			state,
		)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to resolve account")
	})
}
