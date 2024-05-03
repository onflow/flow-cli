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

package super

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flowkit/v2/output"
)

func TestGenerateNewContract(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	_, err := generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile("cadence/tests/TestContract_test.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, testContent)

	// Check content is correct
	expectedContent := `access(all)
contract TestContract {
    init() {}
}`

	expectedTestContent := `import Test

access(all) let account = Test.createAccount()

access(all) fun testContract() {
    let err = Test.deployContract(
        name: "TestContract",
        path: "../contracts/TestContract.cdc",
        arguments: [],
    )

    Test.expect(err, Test.beNil())
}`

	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(fileContent)))
	assert.Equal(t, expectedTestContent, util.NormalizeLineEndings(string(testContent)))

	// Test file already exists scenario
	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.Error(t, err)
	expectedError := fmt.Sprintf("file already exists: %s", filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.Equal(t, expectedError, err.Error())
}

func TestGenerateNewContractSkipTests(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	generateFlags.SkipTests = true

	t.Cleanup(func() {
		generateFlags.SkipTests = false
	})

	// Test contract generation
	_, err := generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile("cadence/tests/TestContract_test.cdc")
	assert.Error(t, err, "Failed to read generated file")
	assert.Nil(t, testContent)
}

func TestGenerateNewContractWithCDCExtension(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	_, err := generateNew([]string{"Tester.cdc"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile("cadence/contracts/Tester.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile("cadence/tests/Tester_test.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, testContent)
}

func TestGenerateNewContractFileAlreadyExists(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	_, err := generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	//// Check if the file exists in the correct directory
	content, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	// Test file already exists scenario
	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.Error(t, err)
	expectedError := fmt.Sprintf("file already exists: %s", filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.Equal(t, expectedError, err.Error())
}

func TestGenerateNewContractWithFileExtension(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	_, err := generateNew([]string{"TestContract.cdc"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	// Check file exists
	content, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)
}

func TestGenerateNewScript(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	_, err := generateNew([]string{"TestScript"}, "script", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile("cadence/scripts/TestScript.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `access(all)
fun main() {
    // Script details here
}`
	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(content)))
}

func TestGenerateNewTransaction(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	_, err := generateNew([]string{"TestTransaction"}, "transaction", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile("cadence/transactions/TestTransaction.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `transaction() {
    prepare(account: &Account) {}

    execute {}
}`
	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(content)))
}

func TestGenerateNewWithDirFlag(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Set a custom directory
	generateFlags.Directory = "customDir"

	_, err := generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile("customDir/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `access(all)
contract TestContract {
    init() {}
}`
	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(content)))
}
