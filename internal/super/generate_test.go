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

	"github.com/onflow/flow-cli/internal/command"

	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/stretchr/testify/assert"
)

func TestGenerateNewInvalid(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	// Mock logger
	logger := output.NewStdoutLogger(output.NoneLog)

	// Test invalid template type
	_, err = generateNew([]string{"invalidType", "TestInvalid"}, command.GlobalFlags{}, logger, nil, nil)
	assert.Error(t, err)
	assert.Equal(t, "invalid template type: invalidType", err.Error())
}

func TestGenerateNewContract(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	logger := output.NewStdoutLogger(output.NoneLog)

	// Test contract generation
	_, err = generateNew([]string{"contract", "TestContract"}, command.GlobalFlags{}, logger, nil, nil)
	assert.NoError(t, err, "Failed to generate contract")

	// Check if the file exists in the correct directory
	assert.FileExists(t, "cadence/contracts/TestContract.cdc")

	content, err := os.ReadFile("cadence/contracts/TestContract.cdc")
	assert.NoError(t, err, "Failed to read generated file")

	// Check content is correct
	expectedContent := `
pub contract TestContract {
    // Contract details here
}`
	assert.Equal(t, expectedContent, string(content))

	// Test file already exists scenario
	_, err = generateNew([]string{"contract", "TestContract"}, command.GlobalFlags{}, logger, nil, nil)
	assert.Error(t, err)
	assert.Equal(t, "file already exists: cadence/contracts/TestContract.cdc", err.Error())
}

func TestGenerateNewContractFileAlreadyExists(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	logger := output.NewStdoutLogger(output.NoneLog)

	// Test contract generation
	_, err = generateNew([]string{"contract", "TestContract"}, command.GlobalFlags{}, logger, nil, nil)
	assert.NoError(t, err, "Failed to generate contract")

	// Check if the file exists in the correct directory
	assert.FileExists(t, "cadence/contracts/TestContract.cdc")

	// Test file already exists scenario
	_, err = generateNew([]string{"contract", "TestContract"}, command.GlobalFlags{}, logger, nil, nil)
	assert.Error(t, err)
	assert.Equal(t, "file already exists: cadence/contracts/TestContract.cdc", err.Error())
}

func TestGenerateNewContractWithFileExtension(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	logger := output.NewStdoutLogger(output.NoneLog)

	_, err = generateNew([]string{"contract", "TestContract.cdc"}, command.GlobalFlags{}, logger, nil, nil)
	assert.NoError(t, err, "Failed to generate contract")

	assert.FileExists(t, "cadence/contracts/TestContract.cdc")
}

func TestGenerateNewScript(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	logger := output.NewStdoutLogger(output.NoneLog)

	_, err = generateNew([]string{"script", "TestScript"}, command.GlobalFlags{}, logger, nil, nil)
	assert.NoError(t, err, "Failed to generate contract")

	assert.FileExists(t, "cadence/scripts/TestScript.cdc")

	content, err := os.ReadFile("cadence/scripts/TestScript.cdc")
	assert.NoError(t, err, "Failed to read generated file")

	expectedContent := `pub fun main() {
    // Script details here
}`
	assert.Equal(t, expectedContent, string(content))
}

func TestGenerateNewTransaction(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(dir)
	os.Chdir(dir)

	logger := output.NewStdoutLogger(output.NoneLog)

	_, err = generateNew([]string{"transaction", "TestTransaction"}, command.GlobalFlags{}, logger, nil, nil)
	assert.NoError(t, err, "Failed to generate contract")

	assert.FileExists(t, "cadence/transactions/TestTransaction.cdc")

	content, err := os.ReadFile("cadence/transactions/TestTransaction.cdc")
	assert.NoError(t, err, "Failed to read generated file")

	expectedContent := `transaction() {
    prepare(account:AuthAccount) {}

    execute {}
}`
	assert.Equal(t, expectedContent, string(content))
}
