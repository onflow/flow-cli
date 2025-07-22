package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	expectedEntries := []string{"# flow", "emulator-account.pkey", ".env", "# Pay attention to imports directory", "!imports/**"}
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

	flowEntries := []string{"# flow", "emulator-account.pkey", ".env", "# Pay attention to imports directory", "!imports/**"}
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
