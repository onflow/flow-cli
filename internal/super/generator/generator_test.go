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

package generator

import (
	"embed"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/util"
)

//go:embed fixtures/*.*
var fixturesFS embed.FS

func TestGenerateNewContract(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)

	// Test contract generation
	err := g.Create(ContractTemplate{Name: "TestContract"})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	// Check content is correct
	expectedContent := `access(all)
contract TestContract {
    init() {}
}`

	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(fileContent)))

	// Test file already exists scenario
	generatorTwo := NewGenerator("", state, logger, false, true)
	err = generatorTwo.Create(ContractTemplate{Name: "TestContract"})
	assert.Error(t, err)
	expectedError := fmt.Sprintf("file already exists: %s", filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.Equal(t, expectedError, err.Error())
}

func TestGenerateContractWithAccount(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)

	// Test contract generation
	err := g.Create(ContractTemplate{Name: "TestContract", Account: "example-account"})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/example-account/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)
}

func TestGenerateNewContractSkipTests(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)

	// Test contract generation
	err := g.Create(ContractTemplate{Name: "TestContract", Account: "", SkipTests: true})
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	testContent, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/tests/TestContract_test.cdc"))
	assert.Error(t, err, "Failed to read generated file")
	assert.Nil(t, testContent)
}

func TestGenerateNewContractFileAlreadyExists(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	g := NewGenerator("", state, logger, false, true)
	err := g.Create(ContractTemplate{Name: "TestContract", Account: ""})
	assert.NoError(t, err, "Failed to generate contract")

	//// Check if the file exists in the correct directory
	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	// Test file already exists scenario
	generatorTwo := NewGenerator("", state, logger, false, true)
	err = generatorTwo.Create(ContractTemplate{Name: "TestContract", Account: ""})
	assert.Error(t, err)
	expectedError := fmt.Sprintf("file already exists: %s", filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.Equal(t, expectedError, err.Error())
}

func TestGenerateNewContractWithFileExtension(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)
	err := g.Create(ContractTemplate{Name: "TestContract.cdc", Account: ""})
	assert.NoError(t, err, "Failed to generate contract")

	// Check file exists
	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)
}

func TestGenerateNewScript(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)
	err := g.Create(ScriptTemplate{Name: "TestScript"})
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

	g := NewGenerator("", state, logger, false, true)
	err := g.Create(TransactionTemplate{Name: "TestTransaction"})
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

	g := NewGenerator("customDir", state, logger, false, true)
	err := g.Create(ContractTemplate{Name: "TestContract", Account: ""})
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("customDir/cadence/contracts/TestContract.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `access(all)
contract TestContract {
    init() {}
}`
	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(content)))
}

func TestGenerateTestTemplate(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)
	err := g.Create(TestTemplate{
		Name:         "Foobar",
		TemplatePath: "contract_init_test.cdc.tmpl",
		Data: map[string]interface{}{
			"ContractName": "Foobar",
		}},
	)
	assert.NoError(t, err, "Failed to generate file")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("cadence/tests/Foobar_test.cdc"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `import Test

access(all) let account = Test.createAccount()

access(all) fun testContract() {
    let err = Test.deployContract(
        name: "Foobar",
        path: "../contracts/Foobar.cdc",
        arguments: [],
    )

    Test.expect(err, Test.beNil())
}`
	assert.Equal(t, expectedContent, util.NormalizeLineEndings(string(content)))
}

func TestGenerateReadmeNoDeps(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)
	err := g.Create(FileTemplate{
		TemplatePath: "README.md.tmpl",
		TargetPath:   "README.md",
		Data: map[string]interface{}{
			"Dependencies": []map[string]interface{}{},
			"Contracts": []map[string]interface{}{
				{"Name": "ExampleContract"},
			},
			"Transactions": []map[string]interface{}{
				{"Name": "ExampleTransaction"},
			},
			"Scripts": []map[string]interface{}{
				{"Name": "ExampleScript"},
			},
			"Tests": []map[string]interface{}{
				{"Name": "ExampleTest"},
			},
		},
	})
	assert.NoError(t, err, "Failed to generate file")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("README.md"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	readmeNoDepsFixture, _ := fixturesFS.ReadFile("fixtures/README_no_deps.md")
	assert.Equal(t, string(readmeNoDepsFixture), string(content))
}

func TestGenerateReadmeWithDeps(t *testing.T) {
	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	g := NewGenerator("", state, logger, false, true)
	err := g.Create(FileTemplate{
		TemplatePath: "README.md.tmpl",
		TargetPath:   "README.md",
		Data: map[string]interface{}{
			"Dependencies": []map[string]interface{}{
				{"Name": "FlowToken"},
				{"Name": "FungibleToken"},
			},
			"Contracts": []map[string]interface{}{
				{"Name": "ExampleContract"},
			},
			"Transactions": []map[string]interface{}{
				{"Name": "ExampleTransaction"},
			},
			"Scripts": []map[string]interface{}{
				{"Name": "ExampleScript"},
			},
			"Tests": []map[string]interface{}{
				{"Name": "ExampleTest"},
			},
		},
	})
	assert.NoError(t, err, "Failed to generate file")

	content, err := state.ReaderWriter().ReadFile(filepath.FromSlash("README.md"))
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	readmeWithDepsFixture, _ := fixturesFS.ReadFile("fixtures/README_with_deps.md")
	assert.Equal(t, string(readmeWithDepsFixture), string(content))
}
