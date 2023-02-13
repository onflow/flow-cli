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
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"os"
	"testing"

	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func TestExecutingTests(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		st, s, _ := services.setup()

		script := tests.TestScriptSimple
		results, err := s.Tests.Execute(script.Source, script.Filename, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()

		st, s, _ := services.setup()

		script := tests.TestScriptSimpleFailing
		results, err := s.Tests.Execute(script.Source, script.Filename, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)

		err = results[0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()

		// Setup
		st, s, _ := services.setup()

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
			Network:  "emulator",
		}
		st.Contracts().AddOrUpdate(c.Name, c)

		// Execute script
		script := tests.TestScriptWithImport
		results, err := s.Tests.Execute(script.Source, script.Filename, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("with file read", func(t *testing.T) {
		t.Parallel()

		// Setup
		st, s, _ := services.setup()
		readerWriter := st.ReaderWriter()
		readerWriter.WriteFile(
			tests.SomeFile.Filename,
			tests.SomeFile.Source,
			os.ModeTemporary,
		)

		// Execute script
		script := tests.TestScriptWithFileRead
		results, err := s.Tests.Execute(script.Source, script.Filename, readerWriter)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})
}
