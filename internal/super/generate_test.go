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

	generator := NewGenerator("", state, logger, false, true)

	// Test contract generation
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: ""}}})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/tests/TestContract_test.cdc"))
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
	generatorTwo := NewGenerator("", state, logger, false, true)
	err = generatorTwo.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: ""}}})
	assert.Error(t, err)
	expectedError := fmt.Sprintf("file already exists: %s", filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.Equal(t, expectedError, err.Error())
}

func TestGenerateContractWithAccount(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	generator := NewGenerator("", state, logger, false, true)

	// Test contract generation
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: "example-account"}}})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/example-account/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/tests/TestContract_test.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, testContent)
}

func TestGenerateNewContractSkipTests(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	generateFlags.SkipTests = true

	generator := NewGenerator("", state, logger, false, true)
	t.Cleanup(func() {
		generateFlags.SkipTests = false
	})

	// Test contract generation
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: ""}}})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/tests/TestContract_test.cdc"))
	assert.Error(t, err, "Failed to read generated file")
	assert.Nil(t, testContent)
}

func TestGenerateNewContractWithCDCExtension(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	generator := NewGenerator("", state, logger, false, true)
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "Tester.cdc", Account: ""}}})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/Tester.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/tests/Tester_test.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, testContent)
}

func TestGenerateNewContractFileAlreadyExists(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	generator := NewGenerator("", state, logger, false, true)
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: ""}}})
	assert.NoError(t, err, "Failed to generate contract")

	//// Check if the file exists in the correct directory
	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	// Test file already exists scenario
	generatorTwo := NewGenerator("", state, logger, false, true)
	err = generatorTwo.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: ""}}})
	assert.Error(t, err)
	expectedError := fmt.Sprintf("file already exists: %s", filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.Equal(t, expectedError, err.Error())
}

func TestGenerateNewContractWithFileExtension(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	generator := NewGenerator("", state, logger, false, true)
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract.cdc", Account: ""}}})
	assert.NoError(t, err, "Failed to generate contract")

	// Check file exists
	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)
}

func TestGenerateNewScript(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	generator := NewGenerator("", state, logger, false, true)
	err := generator.Create(TemplateMap{"script": []TemplateItem{ScriptTemplate{Name: "TestScript"}}})
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/scripts/TestScript.cdc"))
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

	generator := NewGenerator("", state, logger, false, true)
	err := generator.Create(TemplateMap{"transaction": []TemplateItem{TransactionTemplate{Name: "TestTransaction"}}})
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/transactions/TestTransaction.cdc"))
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

	generator := NewGenerator("customDir", state, logger, false, true)
	err := generator.Create(TemplateMap{"contract": []TemplateItem{Contract{Name: "TestContract", Account: ""}}})
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("customDir/cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	testContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("customDir/cadence/tests/TestContract_test.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, testContent)

	expectedContent := `access(all)
contract TestContract {
    init() {}
}`
	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(content)))
}
