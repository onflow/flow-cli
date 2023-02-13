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

	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/tests"
)

func setup() (*flowkit.State, *services.Services, *tests.TestGateway) {
	readerWriter, _ := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	if err != nil {
		panic(err)
	}

	gw := tests.DefaultMockGateway()
	s := services.NewServices(gw.Mock, state, output.NewStdoutLogger(output.NoneLog))

	return state, s, gw
}

func TestExecutingTests(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		st, _, _ := setup()

		script := tests.TestScriptSimple
		results, err := test(script.Source, script.Filename, st, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("simple failing", func(t *testing.T) {
		t.Parallel()

		st, _, _ := setup()

		script := tests.TestScriptSimpleFailing
		results, err := test(script.Source, script.Filename, st, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)

		err = results[0].Error
		require.Error(t, err)
		assert.ErrorAs(t, err, &stdlib.AssertionError{})
	})

	t.Run("with import", func(t *testing.T) {
		t.Parallel()

		// Setup
		st, _, _ := setup()

		c := config.Contract{
			Name:     tests.ContractHelloString.Name,
			Location: tests.ContractHelloString.Filename,
			Network:  "emulator",
		}
		st.Contracts().AddOrUpdate(c.Name, c)

		// Execute script
		script := tests.TestScriptWithImport
		results, err := test(script.Source, script.Filename, st, st.ReaderWriter())

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})

	t.Run("with file read", func(t *testing.T) {
		t.Parallel()

		// Setup
		st, _, _ := setup()
		readerWriter := st.ReaderWriter()
		readerWriter.WriteFile(
			tests.SomeFile.Filename,
			tests.SomeFile.Source,
			os.ModeTemporary,
		)

		// Execute script
		script := tests.TestScriptWithFileRead
		results, err := test(script.Source, script.Filename, st, readerWriter)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.NoError(t, results[0].Error)
	})
}
