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

package util

import (
	"strings"
	"testing"

	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAddressValidForNetwork(t *testing.T) {
	testnetAddr := flow.HexToAddress("8efde57e98c557fa")  // Valid testnet address
	emulatorAddr := flow.HexToAddress("f8d6e0586b0a20c7") // Valid emulator address
	mainnetAddr := flow.HexToAddress("1654653399040a61")  // Valid mainnet address

	t.Run("mainnet address valid for mainnet", func(t *testing.T) {
		assert.True(t, IsAddressValidForNetwork(mainnetAddr, "mainnet"))
	})

	t.Run("mainnet address invalid for testnet", func(t *testing.T) {
		assert.False(t, IsAddressValidForNetwork(mainnetAddr, "testnet"))
	})

	t.Run("testnet address valid for testnet", func(t *testing.T) {
		assert.True(t, IsAddressValidForNetwork(testnetAddr, "testnet"))
	})

	t.Run("testnet address invalid for mainnet", func(t *testing.T) {
		assert.False(t, IsAddressValidForNetwork(testnetAddr, "mainnet"))
	})

	t.Run("emulator address valid for emulator", func(t *testing.T) {
		assert.True(t, IsAddressValidForNetwork(emulatorAddr, "emulator"))
	})

	t.Run("emulator address valid for testing", func(t *testing.T) {
		assert.True(t, IsAddressValidForNetwork(emulatorAddr, "testing"))
	})

	t.Run("custom network allows any address", func(t *testing.T) {
		// Custom networks should allow all addresses since we can't validate without knowing the chain ID
		assert.True(t, IsAddressValidForNetwork(mainnetAddr, "my-custom-network"))
		assert.True(t, IsAddressValidForNetwork(testnetAddr, "my-custom-network"))
		assert.True(t, IsAddressValidForNetwork(emulatorAddr, "my-custom-network"))
	})
}

func TestAddFlowEntriesToGitIgnore_NoDuplicates(t *testing.T) {
	_, state, _ := TestMocks(t)

	err := AddFlowEntriesToGitIgnore("", state.ReaderWriter())
	require.NoError(t, err, "Failed to add Flow entries to gitignore")

	content, err := state.ReaderWriter().ReadFile(".gitignore")
	require.NoError(t, err, "Failed to read gitignore file")

	expectedEntries := []string{"# flow", "emulator-account.pkey", "imports", ".env"}
	for _, entry := range expectedEntries {
		assert.Contains(t, string(content), entry, "Expected gitignore to contain %s", entry)
	}

	err = AddFlowEntriesToGitIgnore("", state.ReaderWriter())
	require.NoError(t, err, "Failed to add Flow entries to gitignore again")

	content, err = state.ReaderWriter().ReadFile(".gitignore")
	require.NoError(t, err, "Failed to read gitignore file again")

	for _, entry := range expectedEntries {
		occurrences := strings.Count(string(content), entry)
		assert.Equal(t, 1, occurrences, "Expected 1 occurrence of %s, but found %d", entry, occurrences)
	}
}

func TestAddFlowEntriesToCursorIgnore_NoDuplicates(t *testing.T) {
	_, state, _ := TestMocks(t)

	err := AddFlowEntriesToCursorIgnore("", state.ReaderWriter())
	require.NoError(t, err, "Failed to add Flow entries to cursorignore")

	content, err := state.ReaderWriter().ReadFile(".cursorignore")
	require.NoError(t, err, "Failed to read cursorignore file")

	expectedEntries := []string{"# flow", "emulator-account.pkey", ".env", "# Pay attention to imports directory", "!imports"}
	for _, entry := range expectedEntries {
		assert.Contains(t, string(content), entry, "Expected cursorignore to contain %s", entry)
	}

	err = AddFlowEntriesToCursorIgnore("", state.ReaderWriter())
	require.NoError(t, err, "Failed to add Flow entries to cursorignore again")

	content, err = state.ReaderWriter().ReadFile(".cursorignore")
	require.NoError(t, err, "Failed to read cursorignore file again")

	for _, entry := range expectedEntries {
		occurrences := strings.Count(string(content), entry)
		assert.Equal(t, 1, occurrences, "Expected 1 occurrence of %s, but found %d", entry, occurrences)
	}
}

func TestAddFlowEntriesToGitIgnore_WithExistingContent(t *testing.T) {
	_, state, _ := TestMocks(t)

	existingContent := "# existing content\nnode_modules/\n*.log\n"
	err := state.ReaderWriter().WriteFile(".gitignore", []byte(existingContent), 0644)
	require.NoError(t, err, "Failed to create existing .gitignore")

	err = AddFlowEntriesToGitIgnore("", state.ReaderWriter())
	require.NoError(t, err, "Failed to add Flow entries to gitignore")

	content, err := state.ReaderWriter().ReadFile(".gitignore")
	require.NoError(t, err, "Failed to read gitignore file")

	assert.Contains(t, string(content), existingContent, "Expected existing content to be preserved")

	flowEntries := []string{"# flow", "emulator-account.pkey", "imports", ".env"}
	for _, entry := range flowEntries {
		assert.Contains(t, string(content), entry, "Expected gitignore to contain %s", entry)
	}
}

func TestAddFlowEntriesToCursorIgnore_WithExistingContent(t *testing.T) {
	_, state, _ := TestMocks(t)

	existingContent := "# existing cursor ignore\n.vscode/\n.idea/\n"
	err := state.ReaderWriter().WriteFile(".cursorignore", []byte(existingContent), 0644)
	require.NoError(t, err, "Failed to create existing .cursorignore")

	err = AddFlowEntriesToCursorIgnore("", state.ReaderWriter())
	require.NoError(t, err, "Failed to add Flow entries to cursorignore")

	content, err := state.ReaderWriter().ReadFile(".cursorignore")
	require.NoError(t, err, "Failed to read cursorignore file")

	assert.Contains(t, string(content), existingContent, "Expected existing content to be preserved")

	flowEntries := []string{"# flow", "emulator-account.pkey", ".env", "# Pay attention to imports directory", "!imports"}
	for _, entry := range flowEntries {
		assert.Contains(t, string(content), entry, "Expected cursorignore to contain %s", entry)
	}
}

func TestAddEntriesToIgnoreFile_HelperFunction(t *testing.T) {
	_, state, _ := TestMocks(t)

	entries := []string{"# test", "test-file.txt", "another-file.log"}
	err := addEntriesToIgnoreFile("test-ignore.txt", entries, state.ReaderWriter())
	require.NoError(t, err, "Failed to add entries to ignore file")

	content, err := state.ReaderWriter().ReadFile("test-ignore.txt")
	require.NoError(t, err, "Failed to read ignore file")

	for _, entry := range entries {
		assert.Contains(t, string(content), entry, "Expected ignore file to contain %s", entry)
	}

	err = addEntriesToIgnoreFile("test-ignore.txt", entries, state.ReaderWriter())
	require.NoError(t, err, "Failed to add entries to ignore file again")

	content, err = state.ReaderWriter().ReadFile("test-ignore.txt")
	require.NoError(t, err, "Failed to read ignore file again")

	for _, entry := range entries {
		occurrences := strings.Count(string(content), entry)
		assert.Equal(t, 1, occurrences, "Expected 1 occurrence of %s, but found %d", entry, occurrences)
	}
}
