/*
 * Flow CLI
 *
 * Copyright 2022 Dapper Labs, Inc.
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

package test

import (
	"os"
	"testing"

	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"

	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutingTests(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		script := tests.TestScriptSimple
		results, err := testCode(script.Source, script.Filename, state)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		script := tests.TestScriptSimpleFailing
		results, err := testCode(script.Source, script.Filename, state)

		require.NoError(t, err)
		require.Len(t, results, 1)

		err = results[0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()
		_, state, _ := util.TestMocks(t)

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
		}
		state.Contracts().AddOrUpdate(c)

		// Execute script
		script := tests.TestScriptWithImport
		results, err := testCode(script.Source, script.Filename, state)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("with file read", func(t *testing.T) {
		t.Parallel()
		_, state, rw := util.TestMocks(t)

		_ = rw.WriteFile(
			tests.SomeFile.Filename,
			tests.SomeFile.Source,
			os.ModeTemporary,
		)

		// Execute script
		script := tests.TestScriptWithFileRead
		results, err := testCode(script.Source, script.Filename, state)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})
}
