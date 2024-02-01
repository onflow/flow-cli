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
	"os"
	"testing"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/flowkit/v2/output"
)

func TestGenerateNewContract(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	fileContent, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, fileContent)

	// Check content is correct
	expectedContent := `
access(all) contract TestContract {
    init() {}
}`
	assert.Equal(t, expectedContent, string(fileContent))

	// Test file already exists scenario
	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.Error(t, err)
	assert.Equal(t, "file already exists: cadence/contracts/TestContract.cdc", err.Error())
}

func TestGenerateNewContractFileAlreadyExists(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Test contract generation
	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	//// Check if the file exists in the correct directory
	content, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	// Test file already exists scenario
	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.Error(t, err)
	assert.Equal(t, "file already exists: cadence/contracts/TestContract.cdc", err.Error())
}

func TestGenerateNewContractWithFileExtension(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	_, err = generateNew([]string{"TestContract.cdc"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	// Check file exists
	content, err := state.ReaderWriter().ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)
}

func TestGenerateNewScript(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	_, err = generateNew([]string{"TestScript"}, "script", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile("cadence/scripts/TestScript.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `access(all) fun main() {
    // Script details here
}`
	assert.Equal(t, expectedContent, string(content))
}

func TestGenerateNewTransaction(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	_, err = generateNew([]string{"TestTransaction"}, "transaction", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile("cadence/transactions/TestTransaction.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `transaction() {
    prepare(account: &Account) {}

    execute {}
}`
	assert.Equal(t, expectedContent, string(content))
}

func TestGenerateNewWithDirFlag(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	logger := output.NewStdoutLogger(output.NoneLog)
	_, state, _ := util.TestMocks(t)

	// Set a custom directory
	generateFlags.Directory = "customDir"

	_, err = generateNew([]string{"TestContract"}, "contract", logger, state)
	assert.NoError(t, err, "Failed to generate contract")

	content, err := state.ReaderWriter().ReadFile("customDir/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")
	assert.NotNil(t, content)

	expectedContent := `
access(all) contract TestContract {
    init() {}
}`
	assert.Equal(t, expectedContent, string(content))
}
